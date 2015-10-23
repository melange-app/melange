var statefulComponent = function(obj) {
    var S = require("./store");
    
    var stateName = "state";
    if (obj.stateName !== undefined) {
        stateName = obj.stateName;
    }

    var unsubscribe = "_unsubscribe";
    var stateUpdate = "_stateUpdate";

    var addFunction = function(name, f) {
        var old = obj[name];
        obj[name] = function() {
            var value = undefined;
            if (old !== undefined) {
                value = old.apply(this, arguments);
            }
            
            return f.apply(this, [value]);
        }
    }

    var select = obj["filterState"];
    if (select == undefined) {
        select = function(s) { return s }
    }

    addFunction("getInitialState", function(state) {
        if (state == undefined) {
            state = {};
        }
        
        state[stateName] = select(S.store.getState());
        return state;
    });

    obj[stateUpdate] = function() {
        var newState = {};
        newState[stateName] = select(S.store.getState());
        this.setState(newState);
    }
    
    addFunction("componentWillMount", function() {
        this[unsubscribe] = S.store.subscribe(this[stateUpdate])
    });
    
    addFunction("componentWillUnmount", function() {
        this[unsubscribe]();
    });
    
    return React.createClass(obj);
}

var classSet = function(obj) {
    var output = "";
    for (var key in obj) {
        if (obj[key]) {
            output += key + " ";
        }
    }
    
    return output;
}

module.exports = {
    createStateful: statefulComponent,
    classSet: classSet,
}
