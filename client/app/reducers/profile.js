var Profile = require('../services/profile');
var Loader = require('./loader');

module.exports = Loader.new("profile", {}, {
    getCurrent: Loader.load(function() {
        var loaded = this.loaded;
        return Profile.current().then(function(p) {
            loaded(p);
        })
    }),
});
