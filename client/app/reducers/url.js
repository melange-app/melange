var actions = {
    update: "_URL__UPDATE",
    back: "_URL__BACK",
    forward: "_URL__FORWARD",
}

var Reducer = function(state, action) {
    if (state == undefined) {
        return Immutable.Map({
            route: undefined,
            history: Immutable.List(),
            forwardHistory: Immutable.List(),
            data: {},
        })
    }
    
    switch (action.type) {
        case actions.update:
            if (action.context.route == state.get('route') &&
                action.context.data == state.get('data')) {
                    return state;
            }

            var history = state.get('history');
            if (state.get('route') !== undefined) {
                history = history.push({
                    route: state.get('route'),
                    data: state.get('data'),
                });
            }
            
            return state.merge({
                route: action.context.route,
                data: action.context.data,
                history: history,
                forwardHistory: Immutable.List(),
            });
            
        case actions.back:
            if (state.get('history').size == 0) {
                return state;
            }
            
            var lastRoute = state.get('history').last();
            
            return state.merge({
                route: lastRoute.route,
                data: lastRoute.data,
                history: state.get('history').pop(),
                forwardHistory: state.get('forwardHistory').unshift({
                    route: state.get('route'),
                    data: state.get('data'),
                })
            });
            
        case actions.forward:
            if (state.get('forwardHistory').size == 0) {
                return state;
            }

            var nextRoute = state.get('forwardHistory').first();

            return state.merge({
                route: nextRoute.route,
                data: nextRoute.data,
                history: state.get('history').push({
                    route: state.get('route'),
                    data: state.get('data'),
                }),
                forwardHistory: state.get('forwardHistory').shift(),
            });
        default:
            return state;
    }
}

module.exports = {
    Reducer: Reducer,
    actions: actions,
}
