const data = require('./data.json')
const fs = require('fs')

function toHump(name) {
    return name.replace(/\_(\w)/g, function(all, letter){
        return letter.toUpperCase();
    });
}

function firstUpperCase(str) {
    return str.replace(/^\S/, function(s){
        return s.toUpperCase();
    });
}


var list = data.data.list
const m = {}
list = list.forEach(it=>{
    const img = it.images[0].img_path || it.images[0].thumb_path
    it.image = `https://dtcg-assets.moecard.cn/img/${img}~card`
    delete it.images
    delete it.card_pack
    delete it.package
    delete it.card_id
    it.cost = parseInt(it.cost) ||0
    it.cost_1 = parseInt(it.cost_1) ||0
    it.DP = parseInt(it.DP) || 0
    it.level = parseInt(it.level) ||0
    const item = {}
    Object.keys(it).forEach(key=>{
        item[firstUpperCase(toHump(key))] = it[key]
    })
    m[it.serial] = item
})
fs.writeFileSync('./data2.json', JSON.stringify(m, null, 2))