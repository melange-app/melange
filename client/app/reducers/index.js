var actions = {};
var registered = {};

var register = function(key, obj) {
    actions[key] = obj.actions;
    registered[key] = obj.Reducer;
}

var Reducer = function(state, action) {
    console.log("ACTION:", action.type);
    
    var obj = {};

    for (var key in registered) {
        var substate = undefined;
        if (state != undefined) {
            substate = state[key];
        }
        
        obj[key] = registered[key](substate, action);
    }

    return obj
}

// State - Simple values about the program state.
register('views', require('./views'));   // views.js, handles newsfeed and menu
register('url', require('./url'));       // urls.js, hanels the routing
register('status', require('./status')); // status.js handles the statusbar

// Stores - Complex data storage.
register('messages', require('./messages')); // messages.js handles message storing
register('plugins', require('./plugins'));   // plugins.js handles plugins
register('identity', require('./identity'));   // identity.js handles identity
register('profile', require('./profile'));   // profile.js handles profiles

module.exports = {
    Reducer: Reducer,
    actions: actions,
};
