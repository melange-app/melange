var actions = {
    loadIdentity: "__PLUGINS_LOAD_IDENTITY",
    setIdentity: "__PLUGINS_SET_IDENTITY",
}

var Identity = require('../services/identity');

var createState = function(current, all, loaded, hasData) {
    var aliases = [];
    if (current !== undefined && all.length > 0) {
        for(var i in all) {
            if (all[i].Fingerprint == current.Fingerprint) {
                aliases = all[i].LoadedAliases;
            }
        }
    }
    
    return {
        loadedAliases: aliases,
        current: current,
        all: all,
        loaded: loaded,
        hasData: hasData,
        load: function(S) {
            S.dispatch(actions.loadIdentity);

            Identity.current().then(function(p) {
                S.dispatch(actions.setIdentity, {
                    loading: false,
                    current: p,
                });
            }).catch(function(err) {
                // Hmm...
                console.log(err);
            });

            Identity.all().then(function(p) {
                S.dispatch(actions.setIdentity, {
                    loading: false,
                    all: p,
                });
            }).catch(function(err) {
                // Hmm...
                console.log(err);
            });
        },
    }
}

var Reducer = function(state, action) {
    var defaultCurrent = undefined;
    var defaultAll = [];
    
    if (state == undefined) {
        return createState(defaultCurrent, defaultAll, false, false);
    }
    
    switch (action.type) {
        case actions.loadIdentity:
            return createState(state.current, state.all, true, state.hasData);
        case actions.setIdentity:
            var current = state.current;
            if (action.context.current !== undefined) {
                current = action.context.current;
            }

            var all = state.all;
            if (action.context.all !== undefined) {
                all = action.context.all;
            }
            
            return createState(current, all, false, true);
        default:
            return state;
    }
}

module.exports = {
    Reducer: Reducer,
    actions: actions,
}
