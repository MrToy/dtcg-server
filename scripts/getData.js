const {execSync} = require('child_process')
const getCmd = `curl 'https://dtcg-api.moecard.cn/api/cdb/cards/search?limit=10000' \
  -H 'authority: dtcg-api.moecard.cn' \
  -H 'sec-ch-ua: " Not A;Brand";v="99", "Chromium";v="96", "Google Chrome";v="96"' \
  -H 'accept: application/json, text/plain, */*' \
  -H 'content-type: application/json' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36' \
  -H 'sec-ch-ua-platform: "macOS"' \
  -H 'origin: https://digimon.card.moe' \
  -H 'sec-fetch-site: cross-site' \
  -H 'sec-fetch-mode: cors' \
  -H 'sec-fetch-dest: empty' \
  -H 'referer: https://digimon.card.moe/' \
  -H 'accept-language: zh-CN,zh;q=0.9' \
  --data-raw '{"keyword":"","language":"ja","card_pack":0,"type":"","color":[],"rarity":[],"tags":[],"tags__logic":"or","order_type":"default","evo_cond":[{}],"qField":[]}' \
  -o data.json`

  execSync(getCmd)