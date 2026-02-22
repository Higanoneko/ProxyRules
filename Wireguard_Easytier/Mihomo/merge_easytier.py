"""
Wireguard_Easytier Mihomo 融合脚本

将 Easytier.js 的 WireGuard 代理增强逻辑注入到
Generator 生成的 mihomo_convert_*.js 转换脚本中，
产出含有 Wireguard_Easytier 加强的转换脚本。

融合策略（链式调用）：
  1. 将原始 main() 重命名为 _originalMain()
  2. 将 Easytier.js 的 main() 重命名为 _easytierEnhance()
  3. 添加新的 main() 串联两者：
     function main(config) { return _easytierEnhance(_originalMain(config)); }
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


def build_bridge_function() -> str:
    """生成桥接函数"""
    return """
// ============ Wireguard_Easytier Bridge ============
function main(config) {
    return _easytierEnhance(_originalMain(config));
}
"""


def merge_single_file(convert_js_path: Path, easytier_code: str) -> str:
    """
    融合单个 mihomo_convert_*.js 文件与 Easytier 代码

    Args:
        convert_js_path: 原始 mihomo_convert_*.js 文件路径
        easytier_code: 已处理的 Easytier 代码（main 已重命名为 _easytierEnhance）

    Returns:
        融合后的 JS 脚本内容
    """
    convert_code = read_file(convert_js_path)

    # 1. 将原始 main() 重命名为 _originalMain()
    modified_convert = rename_main_function(convert_code, "_originalMain")

    # 2. 拼接：原始代码 + Easytier 代码 + 桥接函数
    merged = modified_convert.rstrip()
    merged += "\n\n"
    merged += "// ============ Wireguard_Easytier Start ============\n"
    merged += easytier_code.strip()
    merged += "\n// ============ Wireguard_Easytier End ============\n"
    merged += build_bridge_function()

    return merged


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

    # 读取并处理 Easytier.js —— 将 main 重命名为 _easytierEnhance
    easytier_raw = read_file(EASYTIER_JS)
    easytier_code = rename_main_function(easytier_raw, "_easytierEnhance")

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
        merged_content = merge_single_file(convert_file, easytier_code)
        output_path = OUTPUT_DIR / convert_file.name
        write_file(output_path, merged_content)
        print(f"  [OK] {convert_file.name} -> {output_path.relative_to(PROJECT_ROOT)}")

    print(f"\nDone! {len(convert_files)} enhanced script(s) generated.")


if __name__ == "__main__":
    main()
