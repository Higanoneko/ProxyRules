// Generated at (UTC): 2026-03-24T11:51:13Z

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
const DNS_BOOTSTRAP_LIST = ["119.29.29.29","1.1.1.1","8.8.8.8"];
const DNS_TEMPLATE = {"enable":true,"enhanced-mode":"fake-ip","nameserver":["https://doh.pub/dns-query","https://cloudflare-dns.com/dns-query","https://dns.google/dns-query"],"proxy-server-nameserver":["https://doh.pub/dns-query"],"direct-nameserver":["https://doh.pub/dns-query"],"fake-ip-filter":["*.lan","*.localdomain","*.example","*.invalid","*.localhost","*.test","*.local","*.home.arpa","time.*.com","time.*.gov","time.*.edu.cn","time.*.apple.com","time1.*.com","time2.*.com","time3.*.com","time4.*.com","time5.*.com","time6.*.com","time7.*.com","ntp.*.com","ntp1.*.com","ntp2.*.com","ntp3.*.com","ntp4.*.com","ntp5.*.com","ntp6.*.com","ntp7.*.com","*.time.edu.cn","*.ntp.org.cn","+.pool.ntp.org","time1.cloud.tencent.com","music.163.com","*.music.163.com","*.126.net","*.netease.com","musicapi.taihe.com","music.taihe.com","songsearch.kugou.com","trackercdn.kugou.com","*.kuwo.cn","api-jooxtt.sanook.com","api.joox.com","joox.com","*.tencent.com","*.qq.com","y.qq.com","*.y.qq.com","streamoc.music.tc.qq.com","mobileoc.music.tc.qq.com","isure.stream.qqmusic.qq.com","dl.stream.qqmusic.qq.com","aqqmusic.tc.qq.com","amobile.music.tc.qq.com","*.xiami.com","*.xiaomi.com","*.mi.com","*.music.migu.cn","music.migu.cn","*.msftconnecttest.com","*.msftncsi.com","msftconnecttest.com","msftncsi.com","localhost.ptlogin2.qq.com","localhost.sec.qq.com","+.srv.nintendo.net","+.stun.playstation.net","xbox.*.microsoft.com","+.battlenet.com.cn","+.wotgame.cn","+.wggames.cn","+.wowsgame.cn","+.wargaming.net","proxy.golang.org","stun.*.*","stun.*.*.*","stun.*.*.*.*","heartbeat.belkin.com","*.linksys.com","*.linksyssmartwifi.com","*.router.asus.com","mesu.apple.com","swscan.apple.com","swquery.apple.com","swdownload.apple.com","swcdn.apple.com","swdist.apple.com","lens.l.google.com","stun.l.google.com","+.nflxvideo.net","*.square-enix.com","*.finalfantasyxiv.com","*.ffxiv.com","*.direct","cable.auth.com","network-test.debian.org","detectportal.firefox.com","resolver1.opendns.com","*.xboxlive.com","global.turn.twilio.com","global.stun.twilio.com","app.yinxiang.com","injections.adguard.org","local.adguard.org","localhost.*.qq.com","localhost.*.weixin.qq.com","*.logon.battle.net","*.blzstatic.cn","*.mcdn.bilivideo.cn","*.cmpassport.com","id6.me","open.e.189.cn","opencloud.wostore.cn","id.mail.wo.cn","mdn.open.wo.cn","hmrz.wo.cn","nishub1.10010.com","enrichgw.10010.com","*.wosms.cn","*.jegotrip.com.cn","*.icitymobile.mobi","*.pingan.com.cn","*.cmbchina.com","*.10099.com.cn","*.microdone.cn"]};
const MIXED_PORT = 56365;
const FULL_CONFIG_DEFAULTS = {"allow-lan":true,"mode":"rule","unified-delay":true,"tcp-concurrent":true,"find-process-mode":"strict","global-client-fingerprint":"chrome","log-level":"info","geodata-loader":"standard","external-controller":":9090","external-ui":"./dashboard","external-ui-url":"https://github.com/Zephyruso/zashboard/archive/refs/heads/gh-pages.zip","disable-keep-alive":true,"profile":{"store-selected":true},"geo-auto-update":true,"geo-update-interval":24};
const RULE_PROVIDERS = {"AI":{"type":"http","behavior":"classical","format":"yaml","interval":86400,"url":"https://raw.githubusercontent.com/dler-io/Rules/main/Clash/Provider/AI Suite.yaml","path":"./ruleset/AI.yaml"},"Telegram":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/telegram.txt","path":"./ruleset/Telegram.list"},"YouTube":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/YouTube/YouTube.list","path":"./ruleset/YouTube.list"},"YouTubeMusic":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/YouTubeMusic/YouTubeMusic.list","path":"./ruleset/YouTubeMusic.list"},"Netflix":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Netflix/Netflix.list","path":"./ruleset/Netflix.list"},"TikTok":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/TikTok/TikTok.list","path":"./ruleset/TikTok.list"},"Spotify":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Spotify/Spotify.list","path":"./ruleset/Spotify.list"},"Steam":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Steam/Steam.list","path":"./ruleset/Steam.list"},"Game":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Game/Game.list","path":"./ruleset/Game.list"},"E-Hentai":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/EHGallery/EHGallery.list","path":"./ruleset/E-Hentai.list"},"PornSite":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/PianCat/CustomProxyRuleset/main/PornSite/PornSite.list","path":"./ruleset/PornSite.list"},"Furrybar":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/PianCat/CustomProxyRuleset/main/Furrybar/Furrybar.list","path":"./ruleset/Furrybar.list"},"Stream_US":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/stream_us.txt","path":"./ruleset/Stream_US.list"},"Stream_TW":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/stream_tw.txt","path":"./ruleset/Stream_TW.list"},"Playhorny":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/PianCat/CustomProxyRuleset/main/Playhorny/Playhorny.list","path":"./ruleset/Playhorny.list"},"Stream_JP":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/stream_jp.txt","path":"./ruleset/Stream_JP.list"},"Stream_Global":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/stream.txt","path":"./ruleset/Stream_Global.list"},"Apple":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Apple/Apple.list","path":"./ruleset/Apple.list"},"Microsoft":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Microsoft/Microsoft.list","path":"./ruleset/Microsoft.list"},"Google":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Google/Google.list","path":"./ruleset/Google.list"},"GoogleFCM":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/GoogleFCM/GoogleFCM.list","path":"./ruleset/GoogleFCM.list"},"SogouPrivacy":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/sogouinput.txt","path":"./ruleset/SogouInput.list"},"ADBlock":{"type":"http","behavior":"domain","format":"mrs","interval":86400,"url":"https://adrules.top/adrules-mihomo.mrs","path":"./ruleset/ADBlock.mrs"},"LocalNetwork":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/non_ip/lan.txt","path":"./ruleset/LocalNetwork.list"},"LocalNetworkIP":{"type":"http","behavior":"classical","format":"text","interval":86400,"url":"https://ruleset.skk.moe/Clash/ip/lan.txt","path":"./ruleset/LocalNetworkIP.list"}};
const BASE_RULES = ["RULE-SET,AI,AI","RULE-SET,Telegram,Telegram","RULE-SET,YouTube,YouTube","RULE-SET,YouTubeMusic,YouTube","RULE-SET,Netflix,Netflix","RULE-SET,TikTok,TikTok","RULE-SET,Spotify,Spotify","RULE-SET,Steam,Steam","RULE-SET,Game,Game","RULE-SET,E-Hentai,E-Hentai","RULE-SET,PornSite,PornSite","RULE-SET,Furrybar,PornSite","RULE-SET,Stream_US,US Media","RULE-SET,Stream_TW,Taiwan Media","RULE-SET,Playhorny,Taiwan Media","RULE-SET,Stream_JP,Japan Media","RULE-SET,Stream_Global,Global Media","RULE-SET,Apple,Apple","RULE-SET,Microsoft,Microsoft","RULE-SET,Google,Google","RULE-SET,GoogleFCM,Google FCM","RULE-SET,SogouPrivacy,Sogou Privacy","RULE-SET,ADBlock,ADBlock","RULE-SET,LocalNetwork,DIRECT","RULE-SET,LocalNetworkIP,DIRECT","GEOIP,CN,直接连接","MATCH,选择代理"];
const SNIFFER_CONFIG = {"skip-domain":["Mijia Cloud","dlg.io.mi.com","+.push.apple.com"],"sniff":{"HTTP":{"override-destination":true,"ports":[80,"8080-8880"]},"QUIC":{"ports":[443,8443]},"TLS":{"ports":[443,8443]}}};
const GEOX_URL = {"geoip":"https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip-lite.dat","geosite":"https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat","mmdb":"https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.metadb","asn":"https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/GeoLite2-ASN.mmdb"};
const COUNTRIES = [{"name":"香港","pattern":"(?i)香港|港|HK|hk|Hong Kong|HongKong|hongkong|🇭🇰","icon_url":"https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Hong_Kong.png"},{"name":"台湾","pattern":"(?i)台|新北|彰化|TW|Taiwan|🇹🇼","icon_url":"https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Taiwan.png"},{"name":"新加坡","pattern":"(?i)新加坡|坡|狮城|SG|Singapore|🇸🇬","icon_url":"https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Singapore.png"},{"name":"日本","pattern":"(?i)日本|川日|东京|大阪|泉日|埼玉|沪日|深日|JP|Japan|🇯🇵","icon_url":"https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Japan.png"},{"name":"美国","pattern":"(?i)美国|美|US|United States|🇺🇸","icon_url":"https://testingcf.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/United_States.png"}];
const POLICY_TEMPLATES = [{"name":"选择代理","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Proxy.png","strategy":"selector"},{"name":"手动选择","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Round_Robin_1.png","strategy":"manual"},{"name":"AI","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/AI.png","strategy":"default"},{"name":"Telegram","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Telegram.png","strategy":"default"},{"name":"YouTube","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/YouTube.png","strategy":"default"},{"name":"Netflix","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Netflix.png","strategy":"default"},{"name":"Spotify","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Spotify.png","strategy":"default"},{"name":"TikTok","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/TikTok.png","strategy":"default"},{"name":"Steam","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Steam.png","strategy":"default"},{"name":"Game","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Game.png","strategy":"default"},{"name":"E-Hentai","icon_url":"https://cdn.jsdelivr.net/gh/PianCat/CustomProxyRuleset@main/Icons/Ehentai.png","strategy":"default"},{"name":"PornSite","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Pornhub.png","strategy":"default"},{"name":"US Media","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/United_States.png","strategy":"media_preferred","preferred_country_group":"美国节点"},{"name":"Taiwan Media","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Taiwan.png","strategy":"media_preferred","preferred_country_group":"台湾节点"},{"name":"Japan Media","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Japan.png","strategy":"media_preferred","preferred_country_group":"日本节点"},{"name":"Global Media","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/DomesticMedia.png","strategy":"default"},{"name":"Apple","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Apple.png","strategy":"direct_first"},{"name":"Microsoft","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Microsoft.png","strategy":"direct_first"},{"name":"Google","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Google_Search.png","strategy":"default"},{"name":"Google FCM","icon_url":"https://cdn.jsdelivr.net/gh/PianCat/CustomProxyRuleset@main/Icons/Firebase.png","strategy":"fixed","fixed_proxies":["Google","直接连接"]},{"name":"Sogou Privacy","icon_url":"https://cdn.jsdelivr.net/gh/PianCat/CustomProxyRuleset@main/Icons/Sougou.png","strategy":"fixed","fixed_proxies":["直接连接","REJECT"]},{"name":"ADBlock","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/AdBlack.png","strategy":"fixed","fixed_proxies":["REJECT-DROP","REJECT","直接连接"]},{"name":"直接连接","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Direct.png","strategy":"fixed","fixed_proxies":["DIRECT","选择代理"]},{"name":"GLOBAL","icon_url":"https://cdn.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Global.png","strategy":"global"}];
const ISP_EXCLUDE_PATTERN = "(?i)家宽|家庭|家庭宽带|商宽|商业宽带|星链|Starlink|落地";

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

// ============================================
// 参数定义区域（可根据需要修改）
// ============================================
const ipv6Enabled = false;
const fullConfig = true;
const countryThreshold = 0;
// ============================================
// 参数定义区域结束
// ============================================

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
