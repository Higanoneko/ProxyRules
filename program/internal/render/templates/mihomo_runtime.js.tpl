/*
PianCat 的 Substore 订阅转换脚本
https://github.com/PianCat/ProxyRules

支持的传入参数：
- ipv6: 启用 IPv6 支持（默认 true）
- full: 输出完整配置（适合纯内核启动，默认 false）
- threshold: 国家节点数量小于该值时不显示分组（默认 0）

注意：DNS 始终使用 FakeIP 模式
*/

const NODE_SUFFIX = "节点";
const DNS_BOOTSTRAP_LIST = __DNS_BOOTSTRAP_LIST__;
const DNS_TEMPLATE = __DNS_TEMPLATE__;
const MIXED_PORT = __MIXED_PORT__;
const FULL_CONFIG_DEFAULTS = __FULL_DEFAULTS__;
const RULE_PROVIDERS = __RULE_PROVIDERS__;
const BASE_RULES = __RULES__;
const SNIFFER_CONFIG = __SNIFFER__;
const GEOX_URL = __GEOX_URL__;
const COUNTRIES = __COUNTRIES__;
const POLICY_TEMPLATES = __POLICY_TEMPLATES__;
const ISP_EXCLUDE_PATTERN = __ISP_EXCLUDE_PATTERN__;

function parseBool(value) {
    if (typeof value === "boolean") return value;
    if (typeof value === "string") {
        return value.toLowerCase() === "true" || value === "1";
    }
    return false;
}

function parseNumber(value, defaultValue = 0) {
    if (value === null || typeof value === "undefined") {
        return defaultValue;
    }
    const parsed = parseInt(value, 10);
    return Number.isNaN(parsed) ? defaultValue : parsed;
}

__PARAMETER_BLOCK__

function stripInlineFlag(pattern) {
    return String(pattern || "").replace(/^\(\?i\)/, "");
}

const countryPatternMap = Object.fromEntries(
    COUNTRIES.map((country) => [
        country.name,
        new RegExp(stripInlineFlag(country.pattern), "i"),
    ])
);

const ispRegex = new RegExp(stripInlineFlag(ISP_EXCLUDE_PATTERN), "i");

function buildCountryInventory(proxies, threshold) {
    const counts = Object.create(null);
    let otherCount = 0;

    for (const proxy of proxies) {
        const name = proxy && proxy.name ? proxy.name : "";
        if (ispRegex.test(name)) {
            continue;
        }

        let matched = false;
        for (const country of COUNTRIES) {
            if (countryPatternMap[country.name].test(name)) {
                counts[country.name] = (counts[country.name] || 0) + 1;
                matched = true;
                break;
            }
        }

        if (!matched) {
            otherCount += 1;
        }
    }

    const countries = COUNTRIES
        .map((country) => ({
            name: country.name,
            count: counts[country.name] || 0,
            meta: country,
        }))
        .filter((country) => country.count > 0 && country.count >= threshold);

    return {
        countries,
        names: countries.map((country) => country.name),
        hasOther: otherCount > 0,
    };
}

function buildList(...elements) {
    return elements.flat().filter((value) => value !== null && typeof value !== "undefined" && value !== false && value !== "");
}

function buildDerivedLists(countryGroupNames, hasOther) {
    const otherGroup = hasOther ? "其他节点" : null;
    return {
        selector: buildList(countryGroupNames, otherGroup, "手动选择", "DIRECT"),
        defaults: buildList("选择代理", countryGroupNames, otherGroup, "手动选择", "直接连接"),
        directFirst: buildList("直接连接", "选择代理", countryGroupNames, otherGroup, "手动选择"),
    };
}

function buildPolicyGroup(template, context) {
    const base = {
        name: template.name,
        icon: template.icon_url,
        type: "select",
    };

    switch (template.strategy) {
        case "selector":
            return { ...base, proxies: context.selector };
        case "manual":
            return { ...base, "include-all": true };
        case "default":
            return { ...base, proxies: context.defaults };
        case "media_preferred":
            if (context.countryGroupSet.has(template.preferred_country_group)) {
                return {
                    ...base,
                    proxies: [template.preferred_country_group, "选择代理", "手动选择", "直接连接"],
                };
            }
            return { ...base, proxies: context.defaults };
        case "direct_first":
            return { ...base, proxies: context.directFirst };
        case "fixed":
            return { ...base, proxies: template.fixed_proxies || [] };
        case "global":
            return { ...base, "include-all": true, proxies: context.defaults };
        default:
            return base;
    }
}

function buildCountryGroups(countryInventory) {
    const groups = countryInventory.countries.map((country) => ({
        name: `${country.name}${NODE_SUFFIX}`,
        icon: country.meta.icon_url,
        "include-all": true,
        filter: country.meta.pattern,
        type: "url-test",
        url: "https://cp.cloudflare.com/generate_204",
        interval: 60,
        tolerance: 20,
        lazy: false,
    }));

    if (countryInventory.hasOther) {
        const excludePatterns = COUNTRIES.map((country) => stripInlineFlag(country.pattern)).filter(Boolean);
        groups.push({
            name: "其他节点",
            icon: "https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Global.png",
            "include-all": true,
            type: "select",
            "exclude-filter": excludePatterns.length > 0 ? `(?i)${excludePatterns.join("|")}` : undefined,
        });
    }

    return groups;
}

function buildDnsConfig(ipv6Enabled) {
    const defaultNameserver = DNS_BOOTSTRAP_LIST.filter((dns) => ipv6Enabled || !String(dns).includes(":"));
    return {
        ...DNS_TEMPLATE,
        ipv6: ipv6Enabled,
        "default-nameserver": defaultNameserver,
    };
}

function main(config) {
    const proxies = config && Array.isArray(config.proxies) ? config.proxies : [];
    const resultConfig = { proxies };

    const countryInventory = buildCountryInventory(proxies, countryThreshold);
    const countryGroupNames = countryInventory.names.map((country) => `${country}${NODE_SUFFIX}`);
    const derived = buildDerivedLists(countryGroupNames, countryInventory.hasOther);
    const context = {
        ...derived,
        countryGroupSet: new Set(countryGroupNames),
    };

    const policyGroups = POLICY_TEMPLATES
        .filter((template) => template.strategy !== "global")
        .map((template) => buildPolicyGroup(template, context));
    const globalGroups = POLICY_TEMPLATES
        .filter((template) => template.strategy === "global")
        .map((template) => buildPolicyGroup(template, context));
    const countryGroups = buildCountryGroups(countryInventory);
    const proxyGroups = [...policyGroups, ...countryGroups, ...globalGroups];

    if (fullConfig) {
        Object.assign(resultConfig, FULL_CONFIG_DEFAULTS, {
            "mixed-port": MIXED_PORT,
            ipv6: ipv6Enabled,
        });
    }

    Object.assign(resultConfig, {
        "proxy-groups": proxyGroups,
        "rule-providers": RULE_PROVIDERS,
        rules: [...BASE_RULES],
        sniffer: SNIFFER_CONFIG,
        dns: buildDnsConfig(ipv6Enabled),
        "geodata-mode": true,
        "geox-url": GEOX_URL,
    });

    return resultConfig;
}
