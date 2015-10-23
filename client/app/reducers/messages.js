var actions = {
    loadMessage: "__MESSAGE_LOAD_MESSAGE",
    mergeMessages: "__MESSAGE_MERGE_MESSAGES",
}

var byDate = function(a, b) {
    var dateField = "_parsedDate";

    var aDate = a.get(dateField);
    var bDate = b.get(dateField);

    if (aDate.isBefore(bDate)) {
        return 1;
    } else if (bDate.isBefore(aDate)) {
        return -1;
    }

    return 0;
}

var mergeStore = function(state, messages, messageState) {
    for (var i in messages) {
        var message = messages[i];
        
        // This may become a problem with multiple aliases.
        var id = (message.from.alias + "/" + message.name);
        
        state = updateStore(state, id, message, messageState, false);
    }

    return {
        state: state.state,
        index: state.index,
        store: state.index.sort(byDate),
    }
}

var updateStore = function(state, id, message, messageState, sort) {
    var newStore = state.store;
    var newIndex = state.index;
    
    if (message !== undefined) {
        var obj = {};
        obj[id] = message;
        
        newIndex = newIndex.merge(obj);
    }

    if (sort !== false) {
        newStore = newIndex.sort(byDate);
    }

    var newState = state.state;
    if (messageState) {
        var msgState = {};
        msgState[id] = messageState;

        newState = newState.merge(msgState);
    }
    
    return {
        state: newState,
        index: newIndex,
        store: newStore,
    }
}

var Reducer = function(state, action) {
    if (state == undefined) {
        return {
            state: Immutable.Map(),
            index: Immutable.Map(),
            store: Immutable.List(),
        }
    }
    
    switch (action.type) {
        case actions.loadMessage:
            return updateStore(state,
                               action.context.id,
                               action.context.message,
                               action.context.state);
        case actions.mergeMessages:
            return mergeStore(state, action.context, {
                loaded: true,
            });
        default:
            return state;
    }
}

module.exports = {
    Reducer: Reducer,
    actions: actions,
}
