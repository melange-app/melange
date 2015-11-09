var S = require('../store');
var Components = require('../components');
var Backend = require('../services/backend');
var API = Backend.api;
var UUID = require('../lib/uuid');
console.log(Backend);

var mlgStatus = {
    id: "00000",
    error: {
        code: 0,
        message: "",
    }
}

// Enum for all possible plugin calls.
var messages = {
    pluginInit: "pluginInit",
    viewerUpdate: "viewerUpdate",
    viewerMessage: "viewerMessage",
    findMessages: "findMessages",
    foundMessages: "foundMessages",
    createMessage: "createMessage",
    createdMessage: "createdMessage",
}

var getPluginFromOrigin = function(origin) {
    var testChar = "|";
    var splitChar = encodeURI("|");
    
    var test = API.endpoints.plugins(testChar).split(splitChar);
    
    var suffixChop = test[1].length + test[0].length;
    var prefixChop = test[0].length;
    
    return origin.substr(prefixChop, origin.length - suffixChop);
}

var checkPermissions = function(granted, required, fields) {
    if (granted.has(required)) {
        var grantedFields = granted.get(required).toSet();

        return Immutable.Set(fields).map(function(v) {
            if (v[0] == "?") {
                return v.substr(1, v.length - 1);
            }

            return v;
        }).isSubset(grantedFields);
    }

    return false;
}

var getMessages = function(store, fields) {
    store = store.store;
    
    var setFields = Immutable.Set(fields);
    
    var checkFields = setFields.filterNot(function(key) {
        return key[0] == "?";
    });

    var mapFields = setFields.map(function(key) {
        if (key[0] == "?") {
            return key.substr(1, key.length - 1);
        }

        return key;
    });

    return store.filter(function(m) {
        var hasFields = Immutable.Set.fromKeys(m.get('components'));

        return checkFields.isSubset(hasFields);
    }).map(function(m) {
        var obj = Immutable.fromJS(m).delete('_parsedDate')
        
        return obj.merge({
            components: obj.get('components').filter(function (val, key) {
                return mapFields.has(key);
            }),
        });
    }).toJS();
}

var Plugin = Components.createStateful({
    stateName: "messages",
    filterState: function(s) {
        return s.messages;
    },
    getInitialState: function() {
        return {
            uuid: UUID.generate(),
        }
    },
    componentWillUpdate: function() {
        console.log("Updating...")
    },
    sendMessage: function(type, context) {
        if (type === undefined) {
            console.log("Attempted to send undefined type message.");
            return;
        }
        
        this.refs.iframe.getDOMNode().contentWindow.postMessage({
            type: type,
            context: context,
        }, API.endpoints.plugins(this.props.plugin.get('id')));
    },
    handleMessage: function(event) {
        var plugin = this.props.plugin;
        
        // This message is not for us.
        if (getPluginFromOrigin(event.origin) !== plugin.get('id')) {
            return;
        }
        
        var message = event.data;

        // The id of the plugin doesn't match the ID of the host component.
        if (message.id !== this.state.uuid) {
            return;
        }

        var permissions = plugin.get('permissions');
        if (!(permissions instanceof Immutable.Map)) {
            permissions = Immutable.fromJS(permissions);
        }
        
        switch (message.type) {
            case messages.pluginInit:
                return this.onPluginInit();
            case messages.findMessages:
                return this.onFindMessages(message.context, permissions);
            case messages.viewerUpdate:
                return this.onViewerUpdate(message.context);
            case messages.createMessage:
                return this.onCreateMessage(message.context, permissions);
            default:
                console.log("Cannot handle message type", message.type, "from plugin", plugin.get('id'));
        }
    },
    onViewerUpdate: function(data) {
        if (data.height !== undefined) {
            this.refs.iframe.getDOMNode().height = (data.height + "px");
        }

        if (data.sendMsg) {
            var msg = Immutable.fromJS(this.props.initialMessage).filter(function(v, k) {
                return k.indexOf("_") != 0;
            }).toJS();
            
            this.sendMessage(
                messages.viewerMessage,
                msg
            );
        }
    },
    onFindMessages: function(data, permissions) {
        var hasPermission = checkPermissions(
            permissions,
            "read-message",
            data.fields
        );

        // Plugin had permissions error.
        if (!hasPermission) {
            console.log("Permissions error...");
            return;
        }

        var loaded = getMessages(
            this.state.messages,
            data.fields
        );
        
        this.sendMessage(messages.foundMessages, loaded);

        return;
    },
    onCreateMessage: function(data, permissions) {
        var hasPermission = checkPermissions(
            permissions,
            "send-message",
            data.fields
        );

        // Plugin had permissions error.
        if (!hasPermission) {
            console.log("Permissions error...");
            return;
        }

        // Publish the message
        var component = this;
        Backend.messages._publish(data).then(function() {
            component.sendMessage(
                messages.createdMessage,
                mlgStatus
            );
        })
    },
    onPluginInit: function() {
        console.log("Plugin loaded and ready.");
    },
    componentDidMount: function() {
        window.addEventListener("message", this.handleMessage, false);
    },
    componentWillUnmount: function() {
        window.removeEventListener("message", this.handleMessage, false);
    },
    shouldComponentUpdate: function(nextProps, nextState) {
        return nextProps.plugin.get('id') !== this.props.plugin.get('id');
    },
    render: function() {
        var plugin = this.props.plugin;
        var pluginUrl = API.endpoints.plugins(plugin.get('id')) + "/"
                      + this.props.url + "#id=" + this.state.uuid;
        
        return (
            <iframe sandbox="allow-same-origin allow-scripts"
                    className="plugin"
                    height="20px"
                    src={ pluginUrl }
                    ref="iframe"/>
        );
    }
});

var viewerForMessage = function(msg, plugins) {
    var setComponents = msg.get('components').keySeq();

    var choices = plugins.store.reduce(function(acc, val) {
        var temp = [];

        for (var i in val.viewers) {
            var viewer = val.viewers[i];
            
            temp.push(Immutable.Map({
                id: val.id,
                plugin: Immutable.Map(val),
                view: viewer.view,
                types: Immutable.Set(viewer.type),
                hidden: viewer.hidden,
            }));
        }

        return acc.concat(temp);
    }, Immutable.List()).push(Immutable.Map({
        id: "__melange_profile",
        hidden: false,
        types: Immutable.Set([
            "airdispatch/profile/name",
        ]),
    })).filter(function(p) {
        return p.get('types').isSubset(setComponents);
    });

    if (choices.size == 0) {
        return undefined;
    }

    var viewableChoices = choices.filter(function(p) {
        return !p.get('hidden');
    });

    if (viewableChoices.size == 0) {
        return false;
    }
    
    return viewableChoices.get(0);
}

var isViewable = function(msg, plugins) {
    return (viewerForMessage(msg, plugins) !== false);
}

var Viewer = Components.createStateful({
    stateName: "plugins",
    filterState: function(s) {
        return s.plugins;
    },
    choosePlugin: function() {
        return viewerForMessage(
            this.props.message,
            this.state.plugins
        );
    },
    renderNoPlugin: function() {
        // Print message out.
        return (
            <div className="story-body">
                <p>
                    { this.props.message.get('components').toJS() }
                </p>
            </div>
        );
    },
    renderProfile: function() {
        return (
            <div className="story-body">
                <p>
                    Updated their profile.
                </p>
            </div>
        );
    },
    renderPlugin: function(plugin) {        
        return (
            <div className="story-body viewer">
                <Plugin plugin={ plugin.get('plugin') }
                        url={ plugin.get('view') }
                        initialMessage={ this.props.message }/>
            </div>
        );
    },
    render: function() {
        var plugin = this.choosePlugin();

        if (plugin == undefined) {
            return this.renderNoPlugin();
        }

        if (plugin == false) {
            // This message shouldn't be viewed.

            return (
                <div className="story-body"/>
            );
        }

        if (plugin.get('id') == "__melange_profile") {
            return this.renderProfile();
        }

        return this.renderPlugin(plugin);
    }
});

var Page = React.createClass({
    render: function() {
        return (
            <div className="full-plugin">
            <Plugin plugin={ this.props.data }
                    url="index.html"/>
            </div>
        );
    } 
});

module.exports = {
    Page: Page,
    Plugin: Plugin,
    Viewer: Viewer,
    isViewable: isViewable,
};
