# 赛博电玩 (saibo)

## 店铺信息

| 用途 | URL |
|---|---|
| 淘宝店铺首页 | `https://shop255640440.taobao.com` |
| 店铺分类页 | `https://shop255640440.taobao.com/category.htm` |

赛博电玩是**淘宝店铺**（店铺号 255640440，浙江嘉兴），主营二手 Switch / Switch 2 卡带。经淘宝官方商品页确认：店铺链接为 `https://shop255640440.taobao.com`，商品链接为 `https://pcdetail.taobao.com/<token>.html`（打开需登录）。

## 接口情况

**无公开回收价 API**：与其余六家（老猎人、火枪手、不二家、杭州西子、海螺、蘑菇时间）不同，赛博电玩只有淘宝店铺，没有可匿名访问的 JSON 回收价接口。淘宝页面有登录与反爬（`x5sec`）保护，无法在 `nscardprice` 中拉取价格。

因此本仓库仅注册其商家 key/中文名，`Fetch` 返回 `unavailable`（注明数据来源为淘宝店铺），`sync` / `list` 自动跳过（不支持目录查询）。
