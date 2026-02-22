function main(config) {
    // 1. 确保基础结构存在
    if (!config.proxies) config.proxies = [];
    if (!config.rules) config.rules = [];

    // 2. 定义 Easytier 节点配置
    const easytierProxy = {
        name: "Easytier",
        type: "wireguard",
        server: "<填入 Endpoint 的 IP 或 域名>",
        port: 11013, // 注意：在 JS/YAML 中，端口号必须是数字，不要加引号
        ip: "<填入客户端 Address，例如 10.14.14.2>",
        "public-key": "<填入服务端 PublicKey>",
        "private-key": "<填入客户端 PrivateKey>",
        udp: true
    };

    // 3. 将节点追加到代理列表末尾
    config.proxies.push(easytierProxy);

    // 4. 定义需要插入的分流规则
    const easytierRules = [
        "IP-CIDR,10.19.19.0/24,Easytier,no-resolve",
        "IP-CIDR,10.11.45.0/24,Easytier,no-resolve"
    ];

    // 5. 将 Easytier 规则插入到所有规则的最前面
    config.rules = [...easytierRules, ...config.rules];

    return config;
}