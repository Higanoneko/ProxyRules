/*
Higanoneko 的 Substore 订阅转换脚本
https://github.com/Higanoneko/ProxyRules

支持的传入参数：
- ipv6: 启用 IPv6 支持（默认 true）
- full: 输出完整配置（适合纯内核启动，默认 false）
- dns: 启用 DNS 与域名嗅探（默认 true）
- threshold: 国家节点数量小于该值时不显示分组（默认 0）
*/

const DNS_BOOTSTRAP_LIST = __DNS_BOOTSTRAP_LIST__;
const DNS_TEMPLATE = __DNS_TEMPLATE__;
const MIXED_PORT = __MIXED_PORT__;
const FULL_CONFIG_DEFAULTS = __FULL_DEFAULTS__;
const RULE_PROVIDERS = __RULE_PROVIDERS__;
const BASE_RULES = __RULES__;
const SNIFFER_CONFIG = __SNIFFER__;
const GEOX_URL = __GEOX_URL__;
const POLICY_GROUPS = __POLICY_GROUPS__;
const NODE_EXCLUDE_PATTERN = __NODE_EXCLUDE_PATTERN__;

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

const COUNTRY_GROUPS = POLICY_GROUPS.filter((group) => group.country && group.country !== "其他");
const OTHER_GROUP = POLICY_GROUPS.find((group) => group.country === "其他");
const countryPatternMap = Object.fromEntries(
    COUNTRY_GROUPS.map((group) => [group.name, new RegExp(stripInlineFlag(group.filter), "i")])
);
const excludeRegex = new RegExp(stripInlineFlag(NODE_EXCLUDE_PATTERN), "i");

function buildCountryInventory(proxies, threshold) {
    const counts = Object.create(null);
    let otherCount = 0;

    for (const proxy of proxies) {
        const name = proxy && proxy.name ? proxy.name : "";
        if (excludeRegex.test(name)) {
            continue;
        }

        let matched = false;
        for (const group of COUNTRY_GROUPS) {
            if (countryPatternMap[group.name].test(name)) {
                counts[group.name] = (counts[group.name] || 0) + 1;
                matched = true;
                break;
            }
        }

        if (!matched) {
            otherCount += 1;
        }
    }

    const countries = COUNTRY_GROUPS
        .map((group) => ({
            name: group.name,
            count: counts[group.name] || 0,
        }))
        .filter((country) => country.count > 0 && country.count >= threshold);

    return {
        countries,
        names: countries.map((group) => group.name),
        hasOther: otherCount > 0,
    };
}

function expandProxies(proxies, context) {
    return proxies.flatMap((name) => {
        if (name === "$CountryGroups") return context.countryGroupNames;
        if (context.allCountryGroups.has(name) && !context.availableGroups.has(name)) return [];
        return [name];
    });
}

function buildPolicyGroup(spec, context) {
    const hasMissingCountry = (spec.proxies || []).some((name) =>
        context.allCountryGroups.has(name) && !context.availableGroups.has(name)
    );
    const references = hasMissingCountry && spec.fallback_proxies && spec.fallback_proxies.length
        ? spec.fallback_proxies
        : (spec.proxies || []);
    const proxies = expandProxies(references, context);
    const group = { name: spec.name, icon: spec.icon, type: spec.type };
    if (spec.include_all) group["include-all"] = true;
    if (spec.filter) group.filter = spec.filter;
    if (spec.exclude_filter) {
        group["exclude-filter"] = spec.exclude_filter === "$CountryPatterns"
            ? context.countryExcludePattern
            : spec.exclude_filter;
    }
    if (spec.url) group.url = spec.url;
    if (spec.interval) group.interval = spec.interval;
    if (spec.tolerance) group.tolerance = spec.tolerance;
    if (Object.prototype.hasOwnProperty.call(spec, "lazy")) group.lazy = spec.lazy;
    if (proxies.length) group.proxies = proxies;
    return group;
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
    const countryGroupNames = countryInventory.names.concat(
        countryInventory.hasOther && OTHER_GROUP ? [OTHER_GROUP.name] : []
    );
    const context = {
        countryGroupNames,
        availableGroups: new Set(countryGroupNames),
        allCountryGroups: new Set(POLICY_GROUPS.filter((group) => group.country).map((group) => group.name)),
        countryExcludePattern: `(?i)${COUNTRY_GROUPS.map((group) => stripInlineFlag(group.filter)).join("|")}`,
    };

    const proxyGroups = POLICY_GROUPS
        .filter((spec) => !spec.surge_only && (!spec.country || context.availableGroups.has(spec.name)))
        .map((spec) => buildPolicyGroup(spec, context));

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
        "geodata-mode": true,
        "geox-url": GEOX_URL,
        ...(dnsEnabled ? {
            sniffer: SNIFFER_CONFIG,
            dns: buildDnsConfig(ipv6Enabled),
        } : {}),
    });

    return resultConfig;
}
