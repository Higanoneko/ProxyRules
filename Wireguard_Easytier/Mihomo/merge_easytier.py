"""
Wireguard_Easytier Mihomo 融合脚本

将 Easytier.js 的 WireGuard 代理增强逻辑注入到
Generator 生成的 mihomo_convert_*.js 转换脚本中，
产出含有 Wireguard_Easytier 加强的转换脚本。

融合策略（链式调用 + 配置前置）：
  1. 将 Easytier.js 的用户配置区域（EASYTIER_CONFIG）提取到文件最顶部
  2. 将原始 main() 重命名为 _originalMain()
  3. 将 Easytier.js 的 main() 重命名为 _easytierEnhance()，追加到文件末尾
  4. 添加新的 main() 串联两者
"""

import re
import sys
from pathlib import Path

# 项目根目录
PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent

# 路径定义
EASYTIER_JS = Path(__file__).resolve().parent / "Easytier.js"
SOURCE_DIR = PROJECT_ROOT / "Config" / "Mihomo"
OUTPUT_DIR = Path(__file__).resolve().parent  # 输出到 Wireguard_Easytier/Mihomo/

# 配置区域标记
CONFIG_START = "// ============================================\n// Wireguard_Easytier 用户配置区域"
CONFIG_END = "// Wireguard_Easytier 用户配置区域结束\n// ============================================"


def read_file(path: Path) -> str:
    """读取文件内容"""
    with open(path, "r", encoding="utf-8") as f:
        return f.read()


def write_file(path: Path, content: str):
    """写入文件内容"""
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", encoding="utf-8") as f:
        f.write(content)


def rename_main_function(js_content: str, new_name: str) -> str:
    """
    将 JS 代码中的 function main( 重命名为 function <new_name>(
    只替换函数声明，不影响其他调用。
    """
    return re.sub(
        r'\bfunction\s+main\s*\(',
        f'function {new_name}(',
        js_content,
        count=1  # 只替换第一个匹配
    )


def split_easytier_config_and_logic(easytier_content: str) -> tuple:
    """
    将 Easytier.js 的内容拆分为 用户配置区域 和 逻辑代码 两部分。

    Returns:
        (config_block, logic_block) 元组
    """
    # 找到配置区域的起始和结束
    start_idx = easytier_content.find(CONFIG_START)
    end_idx = easytier_content.find(CONFIG_END)

    if start_idx == -1 or end_idx == -1:
        # 如果找不到标记，整体作为逻辑代码返回
        print("[WARN] Config markers not found in Easytier.js, no config extraction")
        return ("", easytier_content)

    # 配置区域：从标记开始到标记结束（含结束标记）
    config_end_pos = end_idx + len(CONFIG_END)
    config_block = easytier_content[start_idx:config_end_pos].strip()

    # 逻辑代码：配置区域之后的所有内容
    logic_block = easytier_content[config_end_pos:].strip()

    return (config_block, logic_block)


def build_bridge_function() -> str:
    """生成桥接函数"""
    return (
        "// ============ Wireguard_Easytier Bridge ============\n"
        "function main(config) {\n"
        "    return _easytierEnhance(_originalMain(config));\n"
        "}\n"
    )


def insert_config_after_header(convert_code: str, config_block: str) -> str:
    """
    在文件头注释 /* ... */ 之后插入配置区域。
    如果没有找到头注释，则插入到文件最开头。
    """
    if not config_block:
        return convert_code

    # 查找 /* ... */ 块注释的结束位置
    header_end = convert_code.find("*/")
    if header_end != -1:
        # 在 */ 之后插入，保留原有换行
        insert_pos = header_end + 2
        # 跳过紧跟的换行符
        while insert_pos < len(convert_code) and convert_code[insert_pos] in '\r\n':
            insert_pos += 1
        before = convert_code[:insert_pos]
        after = convert_code[insert_pos:]
        return before + "\n" + config_block + "\n\n" + after
    else:
        # 没有头注释，放在最前面
        return config_block + "\n\n" + convert_code


def merge_single_file(convert_js_path: Path, config_block: str, logic_code: str) -> str:
    """
    融合单个 mihomo_convert_*.js 文件与 Easytier 代码

    Args:
        convert_js_path: 原始 mihomo_convert_*.js 文件路径
        config_block: Easytier 用户配置区域
        logic_code: Easytier 逻辑代码（main 已重命名为 _easytierEnhance）

    Returns:
        融合后的 JS 脚本内容
    """
    convert_code = read_file(convert_js_path)

    # 1. 将原始 main() 重命名为 _originalMain()
    modified_convert = rename_main_function(convert_code, "_originalMain")

    # 2. 在文件头注释之后插入配置区域
    modified_convert = insert_config_after_header(modified_convert, config_block)

    # 3. 拼接：原始代码（含配置） + Easytier 逻辑 + 桥接函数
    parts = []

    parts.append(modified_convert.rstrip())
    parts.append("")  # 空行分隔

    # Easytier 逻辑代码（不含配置部分）
    parts.append("// ============ Wireguard_Easytier Start ============")
    parts.append(logic_code.strip())
    parts.append("// ============ Wireguard_Easytier End ============")
    parts.append("")  # 空行分隔

    # 桥接函数
    parts.append(build_bridge_function())

    return "\n".join(parts)


def main():
    """主函数"""
    # 检查 Easytier.js 是否存在
    if not EASYTIER_JS.exists():
        print(f"[ERROR] Easytier.js not found: {EASYTIER_JS}")
        sys.exit(1)

    # 检查源目录是否存在
    if not SOURCE_DIR.exists():
        print(f"[ERROR] Source directory not found: {SOURCE_DIR}")
        print("Please run config_generator.py first to generate mihomo_convert_*.js files.")
        sys.exit(1)

    # 读取 Easytier.js
    easytier_raw = read_file(EASYTIER_JS)

    # 拆分为配置区域和逻辑代码
    config_block, logic_block = split_easytier_config_and_logic(easytier_raw)

    # 将逻辑代码中的 main 重命名为 _easytierEnhance
    logic_code = rename_main_function(logic_block, "_easytierEnhance")

    # 扫描所有 mihomo_convert_*.js 文件
    convert_files = sorted(SOURCE_DIR.glob("mihomo_convert_*.js"))
    if not convert_files:
        print(f"[ERROR] No mihomo_convert_*.js files found in: {SOURCE_DIR}")
        print("Please run config_generator.py first.")
        sys.exit(1)

    print(f"Found {len(convert_files)} convert script(s) to merge.")
    print(f"Easytier source: {EASYTIER_JS}")
    print(f"Output directory: {OUTPUT_DIR}")
    print()

    # 逐个融合
    for convert_file in convert_files:
        merged_content = merge_single_file(convert_file, config_block, logic_code)
        output_path = OUTPUT_DIR / convert_file.name
        write_file(output_path, merged_content)
        print(f"  [OK] {convert_file.name} -> {output_path.relative_to(PROJECT_ROOT)}")

    print(f"\nDone! {len(convert_files)} enhanced script(s) generated.")


if __name__ == "__main__":
    main()
