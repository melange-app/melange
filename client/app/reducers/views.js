var actions = {
    menu: "_VIEWS__MENU",
    newsfeed: "_VIEWS__NEWSFEED",
}

var Reducer = function(state, action) {
    if (state == undefined) {
        return Immutable.Map({
            menu: false,
            newsfeed: false,
        });
    }
    
    switch (action.type) {
    case actions.menu:
        return state.merge({
            menu: !state.get('menu'),
        });
    case actions.newsfeed:
        return state.merge({
            newsfeed: !state.get('newsfeed'),
        });
    default:
        return state;
    }
}

module.exports = {
    Reducer: Reducer,
    actions: actions,
}
