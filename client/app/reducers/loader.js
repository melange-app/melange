var Loader = function(prefix, initial, loaders) {
    var actions = {
        loading: "__" + prefix + "_LOADING",
        loaded: "__" + prefix + "_LOADED",
    }

    var methods = {
        loading: function(S, f) {
            S.dispatch(actions.loading);

            var loaded = function(data) {
                S.dispatch(actions.loaded, {
                    loading: false,
                    data: data,
                });
            }
            
            f.apply({
                loaded: loaded,
            });
        },
    }

    var loaders = Immutable.Map(loaders).map(function(f) {
        return function() {
            return f.apply(methods, arguments);
        }
    });

    var merger = function(old, n) { return n; }
    if (loaders.has("merge")) {
        merger = loaders.get("merge");
        loaders = loaders.delete("merge");
    }

    var createState = function(store, loaded) {
        return loaders.merge({
            store: store,
            loaded: loaded,
        }).toJS();
    }

    var Reducer = function(state, action) {
        if (state == undefined) {
            return createState(initial, false);
        }

        switch (action.type) {
            case actions.loading:
                return createState(state.store, true);
            case actions.loaded:
                return createState(
                    merger(state.store, action.context.data),
                    action.context.loading
                );
            default:
                return state;
        }
    }

    return {
        actions: actions,
        Reducer: Reducer,
    }
}

var Loading = function(f) {
    return function(S) {
        this.loading(S, function() {
            f.apply(this);
        });
    }
}

var LoadingIf = function(f) {
    var loader = Loading(f);
    
    return function(S) {
        if (!this.loaded) {
            return loader(S);
        }

        // do nothing otherwise
    }
}

module.exports = {
    new: Loader,
    load: Loading,
    loadOnce: LoadingIf,
}

