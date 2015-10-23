var Background = function(bgUrl) {
    return {
        "backgroundImage": 'url("' + encodeURI(bgUrl) + '")',
    }
}

module.exports = {
    background: Background,
}
