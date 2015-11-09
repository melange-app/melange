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

var AsyncButton = React.createClass({
    getInitialState: function() {
        return {
            loading: false,
        };
    },
    onClick: function() {
        this.setState({
            loading: true,
        });

        var obj = this;
        this.props.onClick().then(function(res) {
            obj.setState({
                loading: false,
            });

            obj.props.onDone(res);
        });
    },
    renderLoading: function() {
        return (
            <div className="btn btn-primary loading" disabled="disabled">
                { this.props.children }
                <i style={{ "marginLeft": "5px" }} className="fa fa-spinner fa-spin"></i>
            </div>
        );
    },
    renderNormal: function() {
        return (
            <div onClick={ this.onClick } className="btn btn-primary">
                { this.props.children }
            </div>
        );
    },
    render: function() {
        if (this.state.loading) {
            return this.renderLoading();
        } else {
            return this.renderNormal();
        }
        
        return (
            <div className="btn btn-primary">
                { children }
            </div>
        );
    },
});

module.exports = {
    createStateful: statefulComponent,
    classSet: classSet,
    AsyncButton: AsyncButton,
}
