// ============================================
// Wireguard_Easytier 用户配置区域
// 请根据实际情况修改以下配置
// ============================================
const EASYTIER_CONFIG = {
    proxy: {
        name: "Easytier",
        type: "wireguard",
        server: "<填入 Endpoint 的 IP 或 域名>",
        port: 11013,
        ip: "<填入客户端 Address，例如 10.14.14.2>",
        "public-key": "<填入服务端 PublicKey>",
        "private-key": "<填入客户端 PrivateKey>",
        udp: true
    },
    rules: [
        "IP-CIDR,10.19.19.0/24,Easytier,no-resolve",
        "IP-CIDR,10.11.45.0/24,Easytier,no-resolve"
    ]
};
// ============================================
// Wireguard_Easytier 用户配置区域结束
// ============================================

function main(config) {
    // 1. 确保基础结构存在
    if (!config.proxies) config.proxies = [];
    if (!config.rules) config.rules = [];

    // 2. 将节点追加到代理列表末尾
    config.proxies.push(EASYTIER_CONFIG.proxy);

    // 3. 将 Easytier 规则插入到所有规则的最前面
    config.rules = [...EASYTIER_CONFIG.rules, ...config.rules];

    return config;
}