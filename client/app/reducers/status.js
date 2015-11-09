var actions = {
    setStatus: "__STATUS_SET_STATUS",
    setConnection: "__STATUS_SET_CONNECTION",
}

var Reducer = function(state, action) {
    if (state == undefined) {
        return Immutable.Map({
            connectionColor: "yellow",
            connectionText: "Connecting...",
            statusLoading: false,
            statusText: ""
        })
    }
    
    switch (action.type) {
        case actions.setStatus:
            return state.merge({
                statusLoading: action.context.loading,
                statusText: action.context.text,
            });
        case actions.setConnection:
            return state.merge({
                connectionColor: action.context.color,
                connectionText: action.context.text,
            });
        default:
            return state;
    }
}

module.exports = {
    Reducer: Reducer,
    actions: actions,
}
