var actions = {
    loadPlugins: "__PLUGINS_LOAD_PLUGINS",
    setPlugins: "__PLUGINS_SET_PLUGINS",
}

var Plugins = require('../services/plugins');

var createState = function(store, loaded) {
    return {
        store: store,
        loaded: loaded,
        loadAll: function(S) {
            S.dispatch(actions.loadPlugins);

            Plugins.installed().then(function(p) {
                S.dispatch(actions.setPlugins, {
                    loading: false,
                    plugins: p,
                });
            }).catch(function(err) {
                // Hmm...
                console.log(err);
            });
        },
    }
}

var Reducer = function(state, action) {
    if (state == undefined) {
        return createState(Immutable.List(), false);
    }
    
    switch (action.type) {
        case actions.loadPlugins:
            return createState(state.store, true);
        case actions.setPlugins:
            return createState(
                Immutable.List(action.context.plugins),
                action.context.loading
            );
        default:
            return state;
    }
}

module.exports = {
    Reducer: Reducer,
    actions: actions,
}
