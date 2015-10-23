(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
'use strict';

var Status = require('./views/status');
var NewsFeed = require('./views/newsfeed');
var Menu = require('./views/menu');
var Toolbar = require('./views/toolbar');

var Router = require('./router');

var Components = require('./components');

window.images = {
    hunter: "https://scontent-iad3-1.xx.fbcdn.net/hphotos-xpf1/v/t1.0-9/11755276_1598154330449890_3860781029961901461_n.jpg?oh=bf9cfab789484f74fbfe92fb8a3e5d11&oe=56A47885",
    cover: "http://www.smittenblogdesigns.com/wp-content/uploads/2014/01/passion1.png"
};

window.getBackground = function (name) {
    var img = images[name];
    if (name == undefined) {
        img = "/img/icon.png";
    }

    return {
        "backgroundImage": "url('" + img + "')"
    };
};

// Create Redux Store
var S = require("./store");

var Backend = require("./services/backend");

// Export store globally for debugging... (probably a bad idea!)
window.melange = {
    store: S.store,
    backend: Backend
};

var Body = Components.createStateful({
    stateName: "f",
    filterState: function filterState(s) {
        return s.views;
    },
    render: function render() {
        var menuOpen = {
            "body": true,
            "menu-open": this.state.f.get('menu'),
            "newsfeed-open": this.state.f.get('newsfeed')
        };

        return React.createElement(
            'div',
            { className: Components.classSet(menuOpen) },
            React.createElement(NewsFeed, null),
            React.createElement(Menu, null),
            React.createElement(Router, null)
        );
    }
});

var Melange = React.createClass({
    displayName: 'Melange',

    render: function render() {
        return React.createElement(
            'div',
            { className: 'react' },
            React.createElement(Toolbar, null),
            React.createElement(Body, null),
            React.createElement(Status, null)
        );
    }
});

React.render(React.createElement(Melange, null), document.getElementById('melange'), function () {
    console.log("React application started.");

    var loader = document.getElementById("loader");
    loader.style.display = "none";

    Backend.realtime.connect();
});

},{"./components":2,"./router":14,"./services/backend":17,"./store":24,"./views/menu":25,"./views/newsfeed":27,"./views/status":35,"./views/toolbar":36}],2:[function(require,module,exports){
"use strict";

var statefulComponent = function statefulComponent(obj) {
    var S = require("./store");

    var stateName = "state";
    if (obj.stateName !== undefined) {
        stateName = obj.stateName;
    }

    var unsubscribe = "_unsubscribe";
    var stateUpdate = "_stateUpdate";

    var addFunction = function addFunction(name, f) {
        var old = obj[name];
        obj[name] = function () {
            var value = undefined;
            if (old !== undefined) {
                value = old.apply(this, arguments);
            }

            return f.apply(this, [value]);
        };
    };

    var select = obj["filterState"];
    if (select == undefined) {
        select = function (s) {
            return s;
        };
    }

    addFunction("getInitialState", function (state) {
        if (state == undefined) {
            state = {};
        }

        state[stateName] = select(S.store.getState());
        return state;
    });

    obj[stateUpdate] = function () {
        var newState = {};
        newState[stateName] = select(S.store.getState());
        this.setState(newState);
    };

    addFunction("componentWillMount", function () {
        this[unsubscribe] = S.store.subscribe(this[stateUpdate]);
    });

    addFunction("componentWillUnmount", function () {
        this[unsubscribe]();
    });

    return React.createClass(obj);
};

var classSet = function classSet(obj) {
    var output = "";
    for (var key in obj) {
        if (obj[key]) {
            output += key + " ";
        }
    }

    return output;
};

module.exports = {
    createStateful: statefulComponent,
    classSet: classSet
};

},{"./store":24}],3:[function(require,module,exports){
'use strict';

var Background = function Background(bgUrl) {
    return {
        "backgroundImage": 'url("' + encodeURI(bgUrl) + '")'
    };
};

module.exports = {
    background: Background
};

},{}],4:[function(require,module,exports){
/**
 * Fast UUID generator, RFC4122 version 4 compliant.
 * @author Jeff Ward (jcward.com).
 * @license MIT license
 * @link http://stackoverflow.com/questions/105034/how-to-create-a-guid-uuid-in-javascript/21963136#21963136
 **/
'use strict';

var UUID = (function () {
  var self = {};
  var lut = [];for (var i = 0; i < 256; i++) {
    lut[i] = (i < 16 ? '0' : '') + i.toString(16);
  }
  self.generate = function () {
    var d0 = Math.random() * 0xffffffff | 0;
    var d1 = Math.random() * 0xffffffff | 0;
    var d2 = Math.random() * 0xffffffff | 0;
    var d3 = Math.random() * 0xffffffff | 0;
    return lut[d0 & 0xff] + lut[d0 >> 8 & 0xff] + lut[d0 >> 16 & 0xff] + lut[d0 >> 24 & 0xff] + '-' + lut[d1 & 0xff] + lut[d1 >> 8 & 0xff] + '-' + lut[d1 >> 16 & 0x0f | 0x40] + lut[d1 >> 24 & 0xff] + '-' + lut[d2 & 0x3f | 0x80] + lut[d2 >> 8 & 0xff] + '-' + lut[d2 >> 16 & 0xff] + lut[d2 >> 24 & 0xff] + lut[d3 & 0xff] + lut[d3 >> 8 & 0xff] + lut[d3 >> 16 & 0xff] + lut[d3 >> 24 & 0xff];
  };
  return self;
})();

module.exports = UUID;

},{}],5:[function(require,module,exports){
"use strict";

var actions = {
    loadIdentity: "__PLUGINS_LOAD_IDENTITY",
    setIdentity: "__PLUGINS_SET_IDENTITY"
};

var Identity = require('../services/identity');

var createState = function createState(current, all, loaded, hasData) {
    var aliases = [];
    if (current !== undefined && all.length > 0) {
        for (var i in all) {
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
        load: function load(S) {
            S.dispatch(actions.loadIdentity);

            Identity.current().then(function (p) {
                S.dispatch(actions.setIdentity, {
                    loading: false,
                    current: p
                });
            })["catch"](function (err) {
                // Hmm...
                console.log(err);
            });

            Identity.all().then(function (p) {
                S.dispatch(actions.setIdentity, {
                    loading: false,
                    all: p
                });
            })["catch"](function (err) {
                // Hmm...
                console.log(err);
            });
        }
    };
};

var Reducer = function Reducer(state, action) {
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
};

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{"../services/identity":18}],6:[function(require,module,exports){
'use strict';

var actions = {};
var registered = {};

var register = function register(key, obj) {
    actions[key] = obj.actions;
    registered[key] = obj.Reducer;
};

var Reducer = function Reducer(state, action) {
    console.log("ACTION:", action.type);

    var obj = {};

    for (var key in registered) {
        var substate = undefined;
        if (state != undefined) {
            substate = state[key];
        }

        obj[key] = registered[key](substate, action);
    }

    return obj;
};

// State - Simple values about the program state.
register('views', require('./views')); // views.js, handles newsfeed and menu
register('url', require('./url')); // urls.js, hanels the routing
register('status', require('./status')); // status.js handles the statusbar

// Stores - Complex data storage.
register('messages', require('./messages')); // messages.js handles message storing
register('plugins', require('./plugins')); // plugins.js handles plugins
register('identity', require('./identity')); // identity.js handles identity
register('profile', require('./profile')); // profile.js handles profiles

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{"./identity":5,"./messages":8,"./plugins":9,"./profile":10,"./status":11,"./url":12,"./views":13}],7:[function(require,module,exports){
"use strict";

var Loader = function Loader(prefix, initial, loaders) {
    var actions = {
        loading: "__" + prefix + "_LOADING",
        loaded: "__" + prefix + "_LOADED"
    };

    var methods = {
        loading: function loading(S, f) {
            S.dispatch(actions.loading);

            var loaded = function loaded(data) {
                S.dispatch(actions.loaded, {
                    loading: false,
                    data: data
                });
            };

            f.apply({
                loaded: loaded
            });
        }
    };

    var loaders = Immutable.Map(loaders).map(function (f) {
        return function () {
            return f.apply(methods, arguments);
        };
    });

    var merger = function merger(old, n) {
        return n;
    };
    if (loaders.has("merge")) {
        merger = loaders.get("merge");
        loaders = loaders["delete"]("merge");
    }

    var createState = function createState(store, loaded) {
        return loaders.merge({
            store: store,
            loaded: loaded
        }).toJS();
    };

    var Reducer = function Reducer(state, action) {
        if (state == undefined) {
            return createState(initial, false);
        }

        switch (action.type) {
            case actions.loading:
                return createState(state.store, true);
            case actions.loaded:
                return createState(merger(state.store, action.context.data), action.context.loading);
            default:
                return state;
        }
    };

    return {
        actions: actions,
        Reducer: Reducer
    };
};

var Loading = function Loading(f) {
    return function (S) {
        this.loading(S, function () {
            f.apply(this);
        });
    };
};

var LoadingIf = function LoadingIf(f) {
    var loader = Loading(f);

    return function (S) {
        if (!this.loaded) {
            return loader(S);
        }

        // do nothing otherwise
    };
};

module.exports = {
    "new": Loader,
    load: Loading,
    loadOnce: LoadingIf
};

},{}],8:[function(require,module,exports){
"use strict";

var actions = {
    loadMessage: "__MESSAGE_LOAD_MESSAGE",
    mergeMessages: "__MESSAGE_MERGE_MESSAGES"
};

var byDate = function byDate(a, b) {
    var dateField = "_parsedDate";

    var aDate = a.get(dateField);
    var bDate = b.get(dateField);

    if (aDate.isBefore(bDate)) {
        return 1;
    } else if (bDate.isBefore(aDate)) {
        return -1;
    }

    return 0;
};

var mergeStore = function mergeStore(state, messages, messageState) {
    for (var i in messages) {
        var message = messages[i];

        // This may become a problem with multiple aliases.
        var id = message.from.alias + "/" + message.name;

        state = updateStore(state, id, message, messageState, false);
    }

    return {
        state: state.state,
        index: state.index,
        store: state.index.sort(byDate)
    };
};

var updateStore = function updateStore(state, id, message, messageState, sort) {
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
        store: newStore
    };
};

var Reducer = function Reducer(state, action) {
    if (state == undefined) {
        return {
            state: Immutable.Map(),
            index: Immutable.Map(),
            store: Immutable.List()
        };
    }

    switch (action.type) {
        case actions.loadMessage:
            return updateStore(state, action.context.id, action.context.message, action.context.state);
        case actions.mergeMessages:
            return mergeStore(state, action.context, {
                loaded: true
            });
        default:
            return state;
    }
};

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{}],9:[function(require,module,exports){
"use strict";

var actions = {
    loadPlugins: "__PLUGINS_LOAD_PLUGINS",
    setPlugins: "__PLUGINS_SET_PLUGINS"
};

var Plugins = require('../services/plugins');

var createState = function createState(store, loaded) {
    return {
        store: store,
        loaded: loaded,
        loadAll: function loadAll(S) {
            S.dispatch(actions.loadPlugins);

            Plugins.installed().then(function (p) {
                S.dispatch(actions.setPlugins, {
                    loading: false,
                    plugins: p
                });
            })["catch"](function (err) {
                // Hmm...
                console.log(err);
            });
        }
    };
};

var Reducer = function Reducer(state, action) {
    if (state == undefined) {
        return createState(Immutable.List(), false);
    }

    switch (action.type) {
        case actions.loadPlugins:
            return createState(state.store, true);
        case actions.setPlugins:
            return createState(Immutable.List(action.context.plugins), action.context.loading);
        default:
            return state;
    }
};

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{"../services/plugins":21}],10:[function(require,module,exports){
'use strict';

var Profile = require('../services/profile');
var Loader = require('./loader');

module.exports = Loader['new']("profile", {}, {
    getCurrent: Loader.load(function () {
        var loaded = this.loaded;
        return Profile.current().then(function (p) {
            loaded(p);
        });
    })
});

},{"../services/profile":22,"./loader":7}],11:[function(require,module,exports){
"use strict";

var actions = {
    setStatus: "__STATUS_SET_STATUS",
    setConnection: "__STATUS_SET_CONNECTION"
};

var Reducer = function Reducer(state, action) {
    if (state == undefined) {
        return Immutable.Map({
            connectionColor: "yellow",
            connectionText: "Connecting...",
            statusLoading: false,
            statusText: ""
        });
    }

    switch (action.type) {
        case actions.setStatus:
            return state.merge({
                statusLoading: action.context.loading,
                statusText: action.context.text
            });
        case actions.setConnection:
            return state.merge({
                connectionColor: action.context.color,
                connectionText: action.context.text
            });
        default:
            return state;
    }
};

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{}],12:[function(require,module,exports){
"use strict";

var actions = {
    update: "_URL__UPDATE",
    back: "_URL__BACK",
    forward: "_URL__FORWARD"
};

var Reducer = function Reducer(state, action) {
    if (state == undefined) {
        return Immutable.Map({
            route: undefined,
            history: Immutable.List(),
            forwardHistory: Immutable.List(),
            data: {}
        });
    }

    switch (action.type) {
        case actions.update:
            if (action.context.route == state.get('route') && action.context.data == state.get('data')) {
                return state;
            }

            var history = state.get('history');
            if (state.get('route') !== undefined) {
                history = history.push({
                    route: state.get('route'),
                    data: state.get('data')
                });
            }

            return state.merge({
                route: action.context.route,
                data: action.context.data,
                history: history,
                forwardHistory: Immutable.List()
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
                    data: state.get('data')
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
                    data: state.get('data')
                }),
                forwardHistory: state.get('forwardHistory').shift()
            });
        default:
            return state;
    }
};

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{}],13:[function(require,module,exports){
"use strict";

var actions = {
    menu: "_VIEWS__MENU",
    newsfeed: "_VIEWS__NEWSFEED"
};

var Reducer = function Reducer(state, action) {
    if (state == undefined) {
        return Immutable.Map({
            menu: false,
            newsfeed: false
        });
    }

    switch (action.type) {
        case actions.menu:
            return state.merge({
                menu: !state.get('menu')
            });
        case actions.newsfeed:
            return state.merge({
                newsfeed: !state.get('newsfeed')
            });
        default:
            return state;
    }
};

module.exports = {
    Reducer: Reducer,
    actions: actions
};

},{}],14:[function(require,module,exports){
'use strict';

var Components = require('./components');
var Routes = require('./routes');

var Router = Components.createStateful({
    stateName: "url",
    filterState: function filterState(s) {
        return s.url;
    },
    selectPage: function selectPage(url) {
        var current = url.get('route');
        var route = Routes[current];

        // Use default route for the beginning, go to
        // notFound if we cannot load a route.
        if (current == undefined) {
            route = Routes['default'];
        } else if (route == undefined) {
            route = Routes.notFound;
        }

        return route.page;
    },
    render: function render() {
        var page = React.createElement(this.selectPage(this.state.url), {
            data: this.state.url.get('data'),
            route: this.state.url.get('route')
        });

        return React.createElement(
            'div',
            { className: 'page' },
            page
        );
    }
});

module.exports = Router;

},{"./components":2,"./routes":15}],15:[function(require,module,exports){
"use strict";

var routes = {
    home: {
        name: "home",
        page: require('./views/pages/home')
    },
    profile: {
        name: "profile",
        page: require('./views/pages/profile')
    },
    setup: {
        name: "setup",
        page: require('./views/pages/setup')
    },
    settings: {
        name: "settings",
        page: require('./views/pages/settings')
    },
    market: {
        name: "market",
        page: require('./views/pages/market')
    },
    appView: {
        name: "appView",
        page: require('./views/plugin').Page
    }
};

routes.notFound = {
    name: "notFound",
    page: require('./views/pages/notFound')
};

routes["default"] = routes.home;

module.exports = routes;

},{"./views/pages/home":28,"./views/pages/market":29,"./views/pages/notFound":30,"./views/pages/profile":31,"./views/pages/settings":32,"./views/pages/setup":33,"./views/plugin":34}],16:[function(require,module,exports){
"use strict";

var createEndpoint = function createEndpoint(prefix) {
    var melangeSuffix = ".local.getmelange.com:7776";

    if (prefix == "ws") {
        return "ws://api.local.getmelange.com:7776/realtime";
    }

    return "http://" + prefix + melangeSuffix;
};

var endpoints = {
    api: createEndpoint("api"),
    common: createEndpoint("common"),
    plugins: function plugins(pluginId) {
        var prefix = encodeURI(pluginId) + ".plugins";
        return createEndpoint(prefix);
    },
    data: function data() {
        var prefix = createEndpoint("data");

        if (arguments.length == 1) {
            return prefix + "/" + arguments[0];
        } else if (arguments.length == 2) {
            return prefix + "/" + arguments[0] + "/" + arguments[1];
        }

        return prefix;
    },
    ws: createEndpoint("ws")
};

var getCanonicalAPILocation = function getCanonicalAPILocation(url) {
    var prefix = endpoints.api;

    if (url instanceof String) {
        return prefix + url;
    } else if (url instanceof Array) {
        for (var i in url) {
            prefix += "/" + url[i];
        }

        return prefix;
    } else if (url instanceof Immutable.List) {
        return url.reduce(function (acc, val) {
            return acc + "/" + val;
        }, prefix);
    }

    return undefined;
};

var call = function call(method, url, data) {
    var deferred = Q.defer();
    var request = superagent;

    url = getCanonicalAPILocation(url);

    if (method == 'get') {
        request = request.get(url).query(data);
    } else if (method == 'post') {
        request = request.post(url).send(data);
    }

    request.type("json").end(function (err, res) {
        if (res.ok) {
            deferred.resolve(res.body);
        } else {
            deferred.reject({
                status: res.status,
                body: res.text
            });
        }
    });

    return deferred.promise;
};

module.exports = {
    get: function get(url, data) {
        return call('get', url, data);
    },
    post: function post(url, data) {
        return call('post', url, data);
    },
    endpoints: endpoints
};

},{}],17:[function(require,module,exports){
"use strict";

var _exports = {};

var registerModule = function registerModule(name, mod) {
    _exports[name] = mod;
};

registerModule("api", require("./api"));
registerModule("realtime", require("./realtime"));
registerModule("network", require("./network"));
registerModule("plugins", require("./plugins"));
registerModule("identity", require("./identity"));
registerModule("messages", require("./messages"));

module.exports = _exports;

},{"./api":16,"./identity":18,"./messages":19,"./network":20,"./plugins":21,"./realtime":23}],18:[function(require,module,exports){
"use strict";

var api = require("./api");
var prefix = Immutable.List(["identity"]);

var identity = {
    create: function create() {
        return api.post(prefix.push("new"));
    },
    current: function current() {
        return api.get(prefix.push("current"));
    },
    setCurrent: function setCurrent(id) {
        return api.post(prefix.push("current"), {
            fingerprint: id.Fingerprint
        });
    },
    all: function all() {
        return api.get(prefix);
    },
    remove: function remove(id) {
        return api.post(prefix.push("remove"));
    }
};

module.exports = identity;

},{"./api":16}],19:[function(require,module,exports){
"use strict";

var api = require("./api");
var prefix = Immutable.List(["messages"]);
var S = require('../store');

var messages = {
    _get: function _get(alias, name) {
        return api.post(prefix.push("get"), {
            alias: alias,
            name: name
        });
    },
    _sentMessages: function _sentMessages() {
        return api.post();
    },
    _getFromUser: function _getFromUser(alias) {
        return api.post(prefix.push("at"));
    },
    _publish: function _publish(data) {
        return api.post(prefix.push("new"), data);
    },
    get: function get(alias, name) {
        var id = alias + "/" + name;

        S.dispatch(S.actions.messages.loadMessage, {
            id: id,
            state: {
                loaded: false
            }
        });

        messages._get(alias, name).then(function (data) {
            S.dispatch(S.actions.messages.loadMessage, {
                id: id,
                message: data,
                state: {
                    loaded: true
                }
            });
        });
    }
};

module.exports = messages;

},{"../store":24,"./api":16}],20:[function(require,module,exports){
"use strict";

var S = require("../store");

var connect = function connect() {
    window.addEventListener('online', function (event) {
        console.log("Navigator is now online.");
    });

    window.addEventListener('offline', function (event) {
        console.log("Navigator is now offline.");
    });
};

connect();
console.log("Navigator online:", navigator.onLine);

module.exports = {};

},{"../store":24}],21:[function(require,module,exports){
"use strict";

var api = require("./api");
var prefix = Immutable.List(["plugins"]);

var plugins = {
    installed: function installed() {
        return api.get(prefix);
    },
    store: function store() {
        return api.get(prefix.push("store"));
    },
    updates: function updates() {
        return api.get(prefix.push("updates"));
    },
    update: function update(plugin) {
        return api.post(prefix.push("update"));
    },
    install: function install(url) {
        return api.post(prefix.push("install"));
    },
    uninstall: function uninstall(plugin) {
        return api.post(prefix.push("uninstall"));
    }
};

module.exports = plugins;

},{"./api":16}],22:[function(require,module,exports){
"use strict";

var api = require("./api");
var prefix = Immutable.List(["profile"]);

var profiles = {
    update: function update(profile) {
        return api.post(prefix.push("update"), profile);
    },
    current: function current() {
        return api.get(prefix.push("current"));
    }
};

module.exports = profiles;

},{"./api":16}],23:[function(require,module,exports){
"use strict";

var S = require("../store");
var api = require("./api");

// The variable that holds the actual websocket.
var socket;

// A map that holds a list of subscribed functions for
// updates from the websocket.
var subscribers = {};

function sendData(type, data) {
    if (socket == undefined) {
        console.log("Attempted to send data to WebSocket" + " before connection was opened.");
        return;
    }

    // Construct a msg object containing the data the server needs to
    // process the message from the chat client.
    var msg = {
        type: type,
        data: data
    };

    // Send the msg object as a JSON-formatted string.
    socket.send(JSON.stringify(msg));
}

var buffer = Immutable.List();
var firstPassDone = false;

var onMessage = function onMessage(event) {
    var msg = JSON.parse(event.data);
    if (msg["type"] == "initDone") {
        console.log("All messages loaded and in position.");

        firstPassDone = true;

        S.dispatch(S.actions.messages.mergeMessages, buffer.toJS());

        S.dispatch(S.actions.status.setStatus, {
            loading: false,
            text: ""
        });

        return;
    } else if (msg["type"] == "message") {
        // Go ahead and parse the date.
        msg.data["_parsedDate"] = moment(msg.data.date);

        if (!firstPassDone) {
            buffer = buffer.push(msg.data);
        } else {
            S.dispatch(S.actions.messages.mergeMessage, msg.data);
        }

        return;
    } else if (msg["type"] == "waitingForSetup") {
        // This is the first message returned on the websocket.
        // Should be ignored.
        return;
    }

    if (msg["type"] in subscribers) {
        var receivers = subscribers[msg["type"]];
        var toRemove = [];

        for (var i in receivers) {
            try {
                var r = receivers[i].callback(msg["data"]);

                if (r === true) {
                    toRemove.push(i);
                }
            } catch (err) {
                console.log("Couldn't get to callback (" + i + ") for " + msg["type"] + ".");
                console.log(err);
                toRemove.push(i);
            }
        }

        // Remove all the subscribers that we were unable to access.
        for (var i in toRemove) {
            delete receivers[toRemove[i]];
        }

        subscribers[msg["type"]] = receivers;
    } else {
        console.log("Got unknown message with type " + msg["type"]);
    }
};

var onClose = function onClose(event) {
    console.log("Websocket disconnected...");
    console.log(event);
};

var connect = function connect() {
    socket = new WebSocket(api.endpoints.ws);

    // After openeing the WebSocket, we tell the server that we are
    // ready to listen for new data.
    socket.onopen = function (event) {
        console.log("Opened realtime connection.");
        sendData("startup", "ok");

        S.dispatch(S.actions.status.setConnection, {
            color: "green",
            text: "Connected"
        });

        S.dispatch(S.actions.status.setStatus, {
            loading: true,
            text: "Syncing messages..."
        });
    };

    socket.onmessage = onMessage;
    socket.onclose = onClose;
    socket.onerror = onClose;
};

module.exports = {
    connect: connect,
    subscribe: function subscribe(callback) {
        subscribers.push({
            callback: callback
        });
    }
};

},{"../store":24,"./api":16}],24:[function(require,module,exports){
"use strict";

var R = require("./reducers/index");
var store = Redux.createStore(R.Reducer);

var dispatch = function dispatch(type, context) {
    melange.store.dispatch({
        type: type,
        context: context
    });
};

var goto = function goto(route, data) {
    // random default string
    var name = "_";
    if (route !== undefined) {
        name = route.name;
    }

    dispatch(R.actions.url.update, {
        route: name,
        data: data
    });
};

module.exports = {
    actions: R.actions,
    dispatch: dispatch,
    goto: goto,
    store: store
};

},{"./reducers/index":6}],25:[function(require,module,exports){
'use strict';

var S = require('../store');
var Routes = require('../routes');
var API = require('../services/api');

var Components = require('../components');
var Images = require('../images');

var App = React.createClass({
    displayName: 'App',

    openApp: function openApp() {
        console.log("Opening", this.props.plugin.id, this.props.plugin.name);

        S.dispatch(S.actions.views.menu);

        S.goto(Routes.appView, this.props.plugin);
    },
    render: function render() {
        var image = this.props.plugin.image;
        if (image == undefined) {
            image = "/img/icon.png";
        }

        var style = {
            "backgroundImage": "url(\"" + encodeURI(image) + "\")"
        };

        return React.createElement(
            'div',
            { onClick: this.openApp, className: 'app' },
            React.createElement('div', { className: 'app-icon', style: style }),
            React.createElement(
                'p',
                null,
                this.props.plugin.name
            )
        );
    }
});

var Profile = React.createClass({
    displayName: 'Profile',

    renderData: function renderData() {
        var current = this.props.identity.current;
        var aliases = this.props.identity.loadedAliases;

        var address = "";
        if (aliases.length > 0) {
            address = aliases[0].Username + "@" + aliases[0].Location;
        }

        var avatar;
        if (this.props.profile.image) {
            var url = API.endpoints.data(this.props.profile.image);
            avatar = React.createElement('div', { className: 'avatar',
                style: Images.background(url) });
        }

        return React.createElement(
            'div',
            null,
            avatar,
            React.createElement(
                'h1',
                null,
                current.Nickname
            ),
            React.createElement(
                'h2',
                null,
                address
            )
        );
    },
    render: function render() {
        var data = React.createElement('div', null);
        if (this.props.identity.hasData && this.props.profile !== undefined) {
            data = this.renderData();
        }

        return React.createElement(
            'div',
            { className: 'menu-profile' },
            React.createElement(
                'h4',
                null,
                'Signed In As'
            ),
            data
        );
    }
});

var Menu = Components.createStateful({
    stateName: "f",
    filterState: function filterState(s) {
        return {
            plugins: s.plugins,
            identity: s.identity,
            profile: s.profile
        };
    },
    gotoSettings: function gotoSettings() {
        S.dispatch(S.actions.views.menu);
        S.goto(Routes.settings);
    },
    gotoMarket: function gotoMarket() {
        S.dispatch(S.actions.views.menu);
        S.goto(Routes.market);
    },
    closeMenu: function closeMenu() {
        S.dispatch(S.actions.views.menu);
    },
    componentWillMount: function componentWillMount() {
        // We kick off a loading of the plugins.
        // We will get updated (through state) if this changes without us.
        this.state.f.plugins.loadAll(S);
        this.state.f.identity.load(S);
        this.state.f.profile.getCurrent(S);
    },
    render: function render() {
        var allApps = this.state.f.plugins.store.filter(function (p) {
            return !p.hideSidebar;
        }).map(function (p) {
            return React.createElement(App, { plugin: p });
        });

        return React.createElement(
            'div',
            { className: 'menu' },
            React.createElement(
                'div',
                { onClick: this.closeMenu, className: 'close-menu' },
                React.createElement('i', { className: 'fa fa-fw fa-times' }),
                'Close Menu'
            ),
            React.createElement(Profile, { identity: this.state.f.identity,
                profile: this.state.f.profile.store }),
            React.createElement(
                'div',
                { className: 'apps' },
                React.createElement(
                    'h4',
                    null,
                    'Apps'
                ),
                React.createElement(
                    'div',
                    { className: 'app-container' },
                    allApps
                )
            ),
            React.createElement(
                'div',
                { className: 'statics' },
                React.createElement(
                    'div',
                    { onClick: this.gotoMarket, className: 'static' },
                    React.createElement(
                        'p',
                        null,
                        React.createElement('i', { className: 'fa fa-fw fa-shopping-cart' }),
                        'Marketplace'
                    )
                ),
                React.createElement(
                    'div',
                    { onClick: this.gotoSettings, className: 'static' },
                    React.createElement(
                        'p',
                        null,
                        React.createElement('i', { className: 'fa fa-fw fa-gears' }),
                        'Settings'
                    )
                )
            )
        );
    }
});

module.exports = Menu;

},{"../components":2,"../images":3,"../routes":15,"../services/api":16,"../store":24}],26:[function(require,module,exports){
'use strict';

var S = require('../store');
var Viewer = require('./plugin').Viewer;

var Message = React.createClass({
    displayName: 'Message',

    gotoProfile: function gotoProfile(e) {
        e.preventDefault();

        var Routes = require('../routes');

        // close the menu and goto
        S.dispatch(S.actions.views.newsfeed);
        S.goto(Routes.profile, {
            id: this.props.message.getIn(["from", "alias"])
        });
    },
    render: function render() {
        var name = this.props.message.getIn(['from', 'name']);
        if (name == "") {
            name = this.props.message.getIn(['from']);
        }

        // TODO: Implement image proxy on server to prevent XSS attacks.
        var avatar = this.props.message.getIn(['from', 'avatar']);
        if (avatar == "") {
            avatar = "/img/icon.png";
        }
        var avatarStyle = {
            "backgroundImage": "url(\"" + encodeURI(avatar) + "\")"
        };

        var time = this.props.message.get("_parsedDate");
        var timeString = time.fromNow();
        if (time.isBefore(moment().subtract(1, 'days'))) {
            timeString = time.calendar();
        }

        return React.createElement(
            'div',
            { className: 'story' },
            React.createElement('div', { onClick: this.gotoProfile,
                className: 'avatar',
                style: avatarStyle }),
            React.createElement(
                'h3',
                null,
                React.createElement(
                    'a',
                    { onClick: this.gotoProfile, href: '' },
                    name
                )
            ),
            React.createElement(
                'p',
                { className: 'muted', title: time.format('MMMM Do YYYY, h:mm a') },
                timeString
            ),
            React.createElement(Viewer, { message: this.props.message }),
            React.createElement(
                'div',
                { className: 'actions' },
                React.createElement(
                    'a',
                    { href: '' },
                    'Like'
                ),
                ' ·',
                React.createElement(
                    'a',
                    { href: '' },
                    'Comment'
                )
            )
        );
    }
});

module.exports = Message;

},{"../routes":15,"../store":24,"./plugin":34}],27:[function(require,module,exports){
'use strict';

var Components = require('../components');
var Message = require("./message");
var Plugin = require('./plugin');

var NewsFeed = Components.createStateful({
    getInitialState: function getInitialState() {
        return {
            numStories: 25
        };
    },
    stateName: "f",
    filterState: function filterState(s) {
        return {
            messages: s.messages,
            plugins: s.plugins
        };
    },
    render: function render() {
        var plugins = this.state.f.plugins;
        var stories = this.state.f.messages.store.filter(function (value) {
            return Plugin.isViewable(value, plugins);
        }).take(this.state.numStories).map(function (value) {
            return React.createElement(Message, { message: value });
        });

        return React.createElement(
            'div',
            { className: 'newsfeed-holder' },
            React.createElement(
                'div',
                { className: 'newsfeed' },
                React.createElement(
                    'div',
                    { className: 'header' },
                    React.createElement(
                        'div',
                        { className: 'pull-right' },
                        React.createElement(
                            'a',
                            { href: '' },
                            React.createElement('i', { className: 'fa fa-arrows-alt' })
                        )
                    ),
                    'Newsfeed'
                ),
                stories
            )
        );
    }
});

module.exports = NewsFeed;

},{"../components":2,"./message":26,"./plugin":34}],28:[function(require,module,exports){
"use strict";

var S = require("../../store");

var RecentLink = React.createClass({
    displayName: "RecentLink",

    handleClick: function handleClick() {
        this.props.onClick();
    },
    render: function render() {
        return React.createElement(
            "div",
            { className: "recent" },
            React.createElement(
                "div",
                { onClick: this.handleClick, className: "inner" },
                React.createElement("div", { className: "icon", style: { "backgroundImage": "url('/img/icon.png')" } }),
                React.createElement(
                    "h1",
                    null,
                    "Chat with Joey B"
                ),
                React.createElement(
                    "p",
                    null,
                    "September 28, 2015"
                )
            )
        );
    }
});

var Home = React.createClass({
    displayName: "Home",

    gotoSettings: function gotoSettings(e) {
        var Routes = require('../../routes');

        S.goto(Routes.settings);
        e.preventDefault();
    },
    gotoRecent: function gotoRecent() {
        S.goto(Routes.recents);
    },
    gotoApps: function gotoApps(e) {
        S.dispatch(S.actions.views.menu);
        e.preventDefault();
    },
    render: function render() {
        var recents = [React.createElement(RecentLink, { onClick: this.gotoRecent }), React.createElement(RecentLink, { onClick: this.gotoRecent }), React.createElement(RecentLink, { onClick: this.gotoRecent }), React.createElement(RecentLink, { onClick: this.gotoRecent })];

        return React.createElement(
            "div",
            { className: "home" },
            React.createElement(
                "div",
                { className: "home-inner" },
                React.createElement(
                    "div",
                    { className: "search" },
                    React.createElement("input", { type: "text", placeholder: "Search Melange..." })
                ),
                React.createElement(
                    "div",
                    { className: "actions" },
                    React.createElement(
                        "div",
                        { className: "pull-right" },
                        React.createElement(
                            "a",
                            { onClick: this.gotoSettings, href: "" },
                            React.createElement("i", { className: "fa fa-fw fa-gears" }),
                            "Settings"
                        )
                    ),
                    React.createElement(
                        "a",
                        { href: "" },
                        React.createElement("i", { className: "fa fa-fw fa-pencil" }),
                        "Post"
                    )
                ),
                React.createElement(
                    "div",
                    { className: "alert alert-danger" },
                    "You need to renew your name. Please do so now."
                ),
                React.createElement(
                    "div",
                    { className: "recents" },
                    recents
                ),
                React.createElement("hr", null),
                React.createElement(
                    "div",
                    { className: "apps" },
                    React.createElement(
                        "h4",
                        null,
                        React.createElement(
                            "a",
                            { onClick: this.gotoApps, href: "" },
                            React.createElement("i", { className: "fa fa-fw fa-list-alt" }),
                            "Apps"
                        )
                    )
                ),
                React.createElement("hr", null),
                React.createElement(
                    "div",
                    { className: "people" },
                    React.createElement(
                        "h4",
                        null,
                        React.createElement(
                            "a",
                            { href: "" },
                            React.createElement("i", { className: "fa fa-fw fa-users" }),
                            "People"
                        )
                    )
                )
            )
        );
    }
});

module.exports = Home;

},{"../../routes":15,"../../store":24}],29:[function(require,module,exports){
'use strict';

var S = require('../../store');

var TopApp = React.createClass({
    displayName: 'TopApp',

    gotoApp: function gotoApp() {
        var Routes = require('../../routes');

        S.goto(Routes.market, {
            id: "com.getmelange.plugin.chat"
        });
    },
    render: function render() {
        return React.createElement(
            'div',
            { className: 'col-sm-6' },
            React.createElement(
                'div',
                { onClick: this.gotoApp, className: 'panel top-app' },
                React.createElement('div', { className: 'app-icon', style: getBackground() }),
                React.createElement(
                    'h3',
                    null,
                    React.createElement(
                        'span',
                        { className: 'text-muted' },
                        this.props.rank
                    ),
                    ' Chat'
                ),
                React.createElement(
                    'p',
                    null,
                    'from Hunter Leath, com.getmelange.plugin.chat'
                )
            )
        );
    }
});

var MarketHome = React.createClass({
    displayName: 'MarketHome',

    gotoApp: function gotoApp(app) {
        return function () {
            var Routes = require('../../routes');

            S.goto(Routes.market, {
                id: "com.getmelange.plugin.chat"
            });
        };
    },
    render: function render() {
        var apps = [React.createElement(TopApp, { rank: '1' }), React.createElement(TopApp, { rank: '2' }), React.createElement(TopApp, { rank: '3' }), React.createElement(TopApp, { rank: '4' })];

        console.log(this.props);

        return React.createElement(
            'div',
            { className: 'market' },
            React.createElement(
                'div',
                { className: 'featured' },
                React.createElement('div', { onClick: this.gotoApp('com.getmelange.plugins.notes'),
                    className: 'app first', style: getBackground("cover") }),
                React.createElement('div', { onClick: this.gotoApp('com.getmelange.plugins.notes'),
                    className: 'app second', style: getBackground("cover") }),
                React.createElement('div', { onClick: this.gotoApp('com.getmelange.plugins.id'),
                    className: 'app third', style: getBackground("cover") })
            ),
            React.createElement(
                'div',
                { className: 'content' },
                React.createElement(
                    'div',
                    { className: 'alert alert-warning update-alert' },
                    'You have ',
                    React.createElement(
                        'b',
                        null,
                        '6'
                    ),
                    ' applications with updates available. Please review them now.'
                ),
                React.createElement(
                    'h3',
                    null,
                    'Top Applications'
                ),
                React.createElement(
                    'div',
                    { className: 'row top' },
                    apps
                )
            )
        );
    }
});

var MarketView = React.createClass({
    displayName: 'MarketView',

    render: function render() {
        return React.createElement(
            'div',
            { className: 'market-view' },
            React.createElement(
                'div',
                { className: 'row' },
                React.createElement(
                    'div',
                    { className: 'col-sm-4' },
                    React.createElement(
                        'div',
                        { className: 'padding-container' },
                        React.createElement(
                            'div',
                            { className: 'icon-container' },
                            React.createElement('div', { className: 'spacer' }),
                            React.createElement('div', { className: 'app-icon', style: getBackground() })
                        )
                    ),
                    React.createElement(
                        'div',
                        { className: 'actions' },
                        React.createElement(
                            'div',
                            { className: 'btn btn-block btn-large btn-success' },
                            'Install App'
                        ),
                        React.createElement(
                            'div',
                            { className: 'list-group info' },
                            React.createElement(
                                'div',
                                { className: 'list-group-item' },
                                React.createElement(
                                    'span',
                                    { className: 'title text-muted' },
                                    'Author:'
                                ),
                                'Hunter Leath'
                            ),
                            React.createElement(
                                'div',
                                { className: 'list-group-item' },
                                React.createElement(
                                    'span',
                                    { className: 'title text-muted' },
                                    'Version:'
                                ),
                                '0.0.1'
                            ),
                            React.createElement(
                                'div',
                                { className: 'list-group-item' },
                                React.createElement(
                                    'span',
                                    { className: 'title text-muted' },
                                    'Last Updated:'
                                ),
                                'February 15, 2015'
                            ),
                            React.createElement(
                                'div',
                                { className: 'list-group-item' },
                                React.createElement(
                                    'span',
                                    { className: 'title text-muted' },
                                    'Id:'
                                ),
                                this.props.app
                            )
                        ),
                        React.createElement(
                            'div',
                            { className: 'row' },
                            React.createElement(
                                'div',
                                { className: 'col-xs-6' },
                                React.createElement(
                                    'div',
                                    { className: 'btn btn-block btn-primary' },
                                    'Blockchain'
                                )
                            ),
                            React.createElement(
                                'div',
                                { className: 'col-xs-6' },
                                React.createElement(
                                    'div',
                                    { className: 'btn btn-block btn-default' },
                                    'Report'
                                )
                            )
                        )
                    )
                ),
                React.createElement(
                    'div',
                    { className: 'col-sm-8' },
                    React.createElement(
                        'h1',
                        null,
                        'Melange Chat'
                    ),
                    React.createElement(
                        'p',
                        { className: 'text-muted' },
                        'Hunter Leath'
                    ),
                    React.createElement('hr', null),
                    React.createElement(
                        'h3',
                        null,
                        'App Description'
                    ),
                    React.createElement(
                        'p',
                        null,
                        'Pellentesque dapibus suscipit ligula. Donec posuere augue in quam.  Etiam vel tortor sodales tellus ultricies commodo. Suspendisse potenti.  Aenean in sem ac leo mollis blandit.  Donec neque quam, dignissim in, mollis nec, sagittis eu, wisi.  Phasellus lacus.  Etiam laoreet quam sed arcu.  Phasellus at dui in ligula mollis ultricies.  Integer placerat tristique nisl.  Praesent augue.  Fusce commodo.  Vestibulum convallis, lorem a tempus semper, dui dui euismod elit, vitae placerat urna tortor vitae lacus.  Nullam libero mauris, consequat quis, varius et, dictum id, arcu.  Mauris mollis tincidunt felis.  Aliquam feugiat tellus ut neque. Nulla facilisis, risus a rhoncus fermentum, tellus tellus lacinia purus, et dictum nunc justo sit amet elit.'
                    )
                )
            )
        );
    }
});

var Market = React.createClass({
    displayName: 'Market',

    render: function render() {
        if (this.props.data !== undefined) {
            if (this.props.data.has('id')) {
                return React.createElement(MarketView, { app: this.props.data.get('id') });
            }
        }

        return React.createElement(MarketHome, null);
    }
});

module.exports = Market;

},{"../../routes":15,"../../store":24}],30:[function(require,module,exports){
"use strict";

var Error = React.createClass({
    displayName: "Error",

    render: function render() {
        return React.createElement(
            "div",
            { className: "error-page" },
            React.createElement(
                "div",
                { className: "obj" },
                React.createElement("i", { className: "big text-danger fa fa-warning" }),
                React.createElement(
                    "h1",
                    null,
                    "Internal Melange Error"
                ),
                React.createElement("br", null),
                React.createElement(
                    "p",
                    null,
                    "Error Message:"
                ),
                React.createElement(
                    "p",
                    { className: "error" },
                    "Attempted to load route ",
                    this.props.route,
                    ", and could not find it."
                ),
                React.createElement("br", null),
                React.createElement(
                    "h3",
                    null,
                    React.createElement(
                        "a",
                        { href: "" },
                        React.createElement("i", { className: "fa fa-fw fa-exclamation-circle" }),
                        "Report this issue."
                    )
                )
            )
        );
    }
});

var NotFound = React.createClass({
    displayName: "NotFound",

    render: function render() {
        return React.createElement(Error, null);
    }
});

module.exports = NotFound;

},{}],31:[function(require,module,exports){
'use strict';

var S = require('../../store');
var Components = require('../../components');
var Images = require('../../images');
var Message = require('../message');

var Backend = require('../../services/backend');
var API = Backend.api;

var Profile = Components.createStateful({
    stateName: "f",
    filterState: function filterState(s) {
        return {
            profile: s.profile,
            messages: s.messages
        };
    },
    componentWillMount: function componentWillMount() {
        if (this.props.data.has('id')) {
            // Download the other user's messages and profile.
            Backend.messages.get(this.props.data.get('id'), "profile");
        } else {
            this.state.f.profile.getCurrent(S);
        }
    },
    renderBackdrop: function renderBackdrop(profile) {
        var backdrop = {
            "backgroundColor": "#1A8BB0"
        };

        if (profile.backdrop) {
            var url = API.endpoints.data(profile.backdrop);
            backdrop = Images.background(url);
        }

        return React.createElement('div', { className: 'backdrop',
            style: backdrop });
    },
    renderAvatar: function renderAvatar(profile) {
        var avatar;
        if (profile.image) {
            var url = API.endpoints.data(profile.image);
            avatar = React.createElement('div', { className: 'avatar',
                style: Images.background(url) });
        }

        return avatar;
    },
    render: function render() {
        var profile = this.state.f.profile.store;
        profile.messages = this.state.f.messages.store.filter(function (val) {
            return val.get('self') && val.get('public');
        });

        if (this.props.data.has('id')) {
            var alias = this.props.data.get('id');
            var messageId = alias + "/profile";

            var profileMessage = this.state.f.messages.index.get(messageId);
            var profileState = this.state.f.messages.state.get(messageId);

            if (this.state.f.messages.index.has(messageId)) {
                var components = profileMessage.get('components');
                var names = {
                    name: "airdispat.ch/profile/name",
                    description: "airdispat.ch/profile/description",
                    image: "airdispat.ch/profile/avatar"
                };

                profile = {
                    name: components.get(names.name).get('string'),
                    description: components.get(names.description).get('string'),
                    image: components.get(names.image).get('string'),
                    messages: this.state.f.messages.store.filter(function (val) {
                        return val.getIn(['from', 'alias']) == alias && val.get('public');
                    })
                };
            } else {
                profile = {
                    name: "Loading..."
                };
            }
        }

        var description;
        if (profile.description) {
            description = profile.description;
        }

        var messages = profile.messages.map(function (val) {
            return React.createElement(Message, { message: val });
        });

        return React.createElement(
            'div',
            { className: 'profile' },
            this.renderBackdrop(profile),
            this.renderAvatar(profile),
            React.createElement(
                'div',
                { className: 'info' },
                React.createElement(
                    'h1',
                    null,
                    profile.name
                ),
                React.createElement(
                    'p',
                    { className: 'muted' },
                    React.createElement(
                        'a',
                        { href: '' },
                        'Edit My Profile'
                    )
                ),
                React.createElement(
                    'p',
                    null,
                    description
                )
            ),
            React.createElement(
                'div',
                { className: 'profile-feed' },
                messages
            )
        );
    }
});

module.exports = Profile;

},{"../../components":2,"../../images":3,"../../services/backend":17,"../../store":24,"../message":26}],32:[function(require,module,exports){
"use strict";

var Components = require('../../components');

var IdentityOptions = React.createClass({
    displayName: "IdentityOptions",

    render: function render() {
        return React.createElement(
            "div",
            { className: "identity-options" },
            React.createElement("hr", null),
            React.createElement(
                "div",
                { className: "inner" },
                React.createElement(
                    "h5",
                    null,
                    "Options"
                )
            )
        );
    }
});

var Identity = Components.createStateful({
    stateName: "f",
    filterState: function filterState(s) {
        return {
            profile: s.profile,
            identity: s.identity
        };
    },
    componentWillMount: function componentWillMount() {
        this.state.f.profile.getCurrent(S);
        this.state.f.identity.load(S);
    },
    getInitialState: function getInitialState() {
        return {
            open: false
        };
    },
    toggleSettings: function toggleSettings() {
        this.setState({
            open: !this.state.open
        });
    },
    render: function render() {
        console.log(this.state);

        var identityClass = "identity";
        if (this.props.active === true) {
            identityClass += " active";
        } else {
            identityClass += " inactive";
        }

        var open = undefined;
        if (this.state.open) {
            open = React.createElement(IdentityOptions, null);
        }

        return React.createElement(
            "div",
            { className: identityClass },
            React.createElement(
                "div",
                { onClick: this.toggleSettings, className: "identity-settings" },
                React.createElement("i", { className: "fa fa-cog" })
            ),
            React.createElement(
                "div",
                { className: "panel" },
                React.createElement("div", { className: "avatar", style: getBackground("hunter") }),
                React.createElement(
                    "div",
                    { className: "go-icon" },
                    React.createElement("i", { className: "fa fa-chevron-right" })
                ),
                React.createElement(
                    "div",
                    { className: "content" },
                    React.createElement(
                        "h4",
                        null,
                        "Hunter Leath"
                    ),
                    React.createElement(
                        "p",
                        { className: "text-muted" },
                        "hleath@airdispat.ch"
                    )
                ),
                open
            )
        );
    }
});

var Settings = React.createClass({
    displayName: "Settings",

    render: function render() {
        var identities = [React.createElement(Identity, null), React.createElement(Identity, null), React.createElement(Identity, null)];

        return React.createElement(
            "div",
            { className: "settings" },
            React.createElement(
                "h1",
                null,
                "My Identities"
            ),
            React.createElement(
                "h3",
                null,
                "Current User"
            ),
            React.createElement(Identity, { active: true }),
            React.createElement(
                "h3",
                null,
                "Switch User ",
                React.createElement(
                    "small",
                    null,
                    "or ",
                    React.createElement(
                        "a",
                        { href: "" },
                        "Create New User"
                    )
                )
            ),
            identities
        );
    }
});

module.exports = Settings;

},{"../../components":2}],33:[function(require,module,exports){
"use strict";

var Setup = React.createClass({
    displayName: "Setup",

    render: function render() {
        return React.createElement(
            "div",
            { className: "setup" },
            React.createElement(
                "p",
                { className: "lead" },
                "Welcome to"
            ),
            React.createElement("br", null),
            React.createElement(
                "div",
                { className: "row" },
                React.createElement(
                    "div",
                    { className: "col-xs-3" },
                    React.createElement("img", { className: "img-responsive", src: "/img/icon.png" })
                ),
                React.createElement(
                    "div",
                    { className: "col-xs-8" },
                    React.createElement(
                        "h1",
                        { className: "style" },
                        "Melange"
                    ),
                    React.createElement(
                        "h3",
                        null,
                        "Developer Preview"
                    )
                )
            ),
            React.createElement("br", null),
            React.createElement(
                "div",
                { className: "row" },
                React.createElement(
                    "div",
                    { className: "col-xs-6" },
                    React.createElement(
                        "p",
                        { className: "lead" },
                        "The secure social network that gives you back control of your data."
                    ),
                    React.createElement(
                        "div",
                        { className: "row text-center" },
                        React.createElement(
                            "div",
                            { className: "col-xs-6" },
                            React.createElement(
                                "h1",
                                null,
                                React.createElement("i", { className: "fa fa-cloud-download" })
                            ),
                            React.createElement(
                                "p",
                                null,
                                "Install Apps"
                            )
                        ),
                        React.createElement(
                            "div",
                            { className: "col-xs-6" },
                            React.createElement(
                                "h1",
                                null,
                                React.createElement("i", { className: "fa fa-users" })
                            ),
                            React.createElement(
                                "p",
                                null,
                                "Find Friends"
                            )
                        )
                    )
                ),
                React.createElement(
                    "div",
                    { className: "col-xs-6" },
                    React.createElement(SetupButton, { text: "Sign Up" }),
                    React.createElement(SetupButton, { text: "Link Existing Account", color: "pink" })
                )
            )
        );
    }
});

var SetupButton = React.createClass({
    displayName: "SetupButton",

    render: function render() {
        var style = {};
        if (this.props.color != undefined) {
            style["border-bottom-color"] = this.props.color;
        }

        return React.createElement(
            "div",
            { className: "button", style: style },
            React.createElement(
                "p",
                null,
                this.props.text,
                " →"
            )
        );
    }
});

module.exports = Setup;

},{}],34:[function(require,module,exports){
'use strict';

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
        message: ""
    }
};

// Enum for all possible plugin calls.
var messages = {
    pluginInit: "pluginInit",
    viewerUpdate: "viewerUpdate",
    viewerMessage: "viewerMessage",
    findMessages: "findMessages",
    foundMessages: "foundMessages",
    createMessage: "createMessage",
    createdMessage: "createdMessage"
};

var getPluginFromOrigin = function getPluginFromOrigin(origin) {
    var testChar = "|";
    var splitChar = encodeURI("|");

    var test = API.endpoints.plugins(testChar).split(splitChar);

    var suffixChop = test[1].length + test[0].length;
    var prefixChop = test[0].length;

    return origin.substr(prefixChop, origin.length - suffixChop);
};

var checkPermissions = function checkPermissions(granted, required, fields) {
    if (granted.has(required)) {
        var grantedFields = granted.get(required).toSet();

        return Immutable.Set(fields).map(function (v) {
            if (v[0] == "?") {
                return v.substr(1, v.length - 1);
            }

            return v;
        }).isSubset(grantedFields);
    }

    return false;
};

var getMessages = function getMessages(store, fields) {
    store = store.store;

    var setFields = Immutable.Set(fields);

    var checkFields = setFields.filterNot(function (key) {
        return key[0] == "?";
    });

    var mapFields = setFields.map(function (key) {
        if (key[0] == "?") {
            return key.substr(1, key.length - 1);
        }

        return key;
    });

    return store.filter(function (m) {
        var hasFields = Immutable.Set.fromKeys(m.components);

        return checkFields.isSubset(hasFields);
    }).map(function (m) {
        var obj = Immutable.fromJS(m)['delete']('_parsedDate');

        return obj.merge({
            components: obj.get('components').filter(function (val, key) {
                return mapFields.has(key);
            })
        });
    }).toJS();
};

var Plugin = Components.createStateful({
    stateName: "messages",
    filterState: function filterState(s) {
        return s.messages;
    },
    getInitialState: function getInitialState() {
        return {
            uuid: UUID.generate()
        };
    },
    componentWillUpdate: function componentWillUpdate() {
        console.log("Updating...");
    },
    sendMessage: function sendMessage(type, context) {
        if (type === undefined) {
            console.log("Attempted to send undefined type message.");
            return;
        }

        this.refs.iframe.getDOMNode().contentWindow.postMessage({
            type: type,
            context: context
        }, API.endpoints.plugins(this.props.plugin.get('id')));
    },
    handleMessage: function handleMessage(event) {
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
    onViewerUpdate: function onViewerUpdate(data) {
        if (data.height !== undefined) {
            this.refs.iframe.getDOMNode().height = data.height + "px";
        }

        if (data.sendMsg) {
            var msg = Immutable.fromJS(this.props.initialMessage).filter(function (v, k) {
                return k.indexOf("_") != 0;
            }).toJS();

            this.sendMessage(messages.viewerMessage, msg);
        }
    },
    onFindMessages: function onFindMessages(data, permissions) {
        var hasPermission = checkPermissions(permissions, "read-message", data.fields);

        // Plugin had permissions error.
        if (!hasPermission) {
            console.log("Permissions error...");
            return;
        }

        var loaded = getMessages(this.state.messages, data.fields);

        this.sendMessage(messages.foundMessages, loaded);

        return;
    },
    onCreateMessage: function onCreateMessage(data, permissions) {
        var hasPermission = checkPermissions(permissions, "send-message", data.fields);

        // Plugin had permissions error.
        if (!hasPermission) {
            console.log("Permissions error...");
            return;
        }

        // Publish the message
        var component = this;
        Backend.messages._publish(data).then(function () {
            component.sendMessage(messages.createdMessage, mlgStatus);
        });
    },
    onPluginInit: function onPluginInit() {
        console.log("Plugin loaded and ready.");
    },
    componentDidMount: function componentDidMount() {
        window.addEventListener("message", this.handleMessage, false);
    },
    componentWillUnmount: function componentWillUnmount() {
        window.removeEventListener("message", this.handleMessage, false);
    },
    shouldComponentUpdate: function shouldComponentUpdate(nextProps, nextState) {
        return nextProps.plugin.get('id') !== this.props.plugin.get('id');
    },
    render: function render() {
        var plugin = this.props.plugin;
        var pluginUrl = API.endpoints.plugins(plugin.get('id')) + "/" + this.props.url + "#id=" + this.state.uuid;

        return React.createElement('iframe', { sandbox: 'allow-same-origin allow-scripts',
            className: 'plugin',
            height: '20px',
            src: pluginUrl,
            ref: 'iframe' });
    }
});

var viewerForMessage = function viewerForMessage(msg, plugins) {
    var setComponents = msg.get('components').keySeq();

    var choices = plugins.store.reduce(function (acc, val) {
        var temp = [];

        for (var i in val.viewers) {
            var viewer = val.viewers[i];

            temp.push(Immutable.Map({
                id: val.id,
                plugin: Immutable.Map(val),
                view: viewer.view,
                types: Immutable.Set(viewer.type),
                hidden: viewer.hidden
            }));
        }

        return acc.concat(temp);
    }, Immutable.List()).push(Immutable.Map({
        id: "__melange_profile",
        hidden: false,
        types: Immutable.Set(["airdispatch/profile/name"])
    })).filter(function (p) {
        return p.get('types').isSubset(setComponents);
    });

    if (choices.size == 0) {
        return undefined;
    }

    var viewableChoices = choices.filter(function (p) {
        return !p.get('hidden');
    });

    if (viewableChoices.size == 0) {
        return false;
    }

    return viewableChoices.get(0);
};

var isViewable = function isViewable(msg, plugins) {
    return viewerForMessage(msg, plugins) !== false;
};

var Viewer = Components.createStateful({
    stateName: "plugins",
    filterState: function filterState(s) {
        return s.plugins;
    },
    choosePlugin: function choosePlugin() {
        return viewerForMessage(this.props.message, this.state.plugins);
    },
    renderNoPlugin: function renderNoPlugin() {
        // Print message out.
        return React.createElement(
            'div',
            { className: 'story-body' },
            React.createElement(
                'p',
                null,
                this.props.message.get('components').toJS()
            )
        );
    },
    renderProfile: function renderProfile() {
        return React.createElement(
            'div',
            { className: 'story-body' },
            React.createElement(
                'p',
                null,
                'Updated their profile.'
            )
        );
    },
    renderPlugin: function renderPlugin(plugin) {
        return React.createElement(
            'div',
            { className: 'story-body viewer' },
            React.createElement(Plugin, { plugin: plugin.get('plugin'),
                url: plugin.get('view'),
                initialMessage: this.props.message })
        );
    },
    render: function render() {
        var plugin = this.choosePlugin();

        if (plugin == undefined) {
            return this.renderNoPlugin();
        }

        if (plugin == false) {
            // This message shouldn't be viewed.

            return React.createElement('div', { className: 'story-body' });
        }

        if (plugin.get('id') == "__melange_profile") {
            return this.renderProfile();
        }

        return this.renderPlugin(plugin);
    }
});

var Page = React.createClass({
    displayName: 'Page',

    render: function render() {
        return React.createElement(
            'div',
            { className: 'full-plugin' },
            React.createElement(Plugin, { plugin: this.props.data,
                url: 'index.html' })
        );
    }
});

module.exports = {
    Page: Page,
    Plugin: Plugin,
    Viewer: Viewer,
    isViewable: isViewable
};

},{"../components":2,"../lib/uuid":4,"../services/backend":17,"../store":24}],35:[function(require,module,exports){
"use strict";

var Components = require('../components');

var Status = Components.createStateful({
    stateName: "f",
    filterState: function filterState(s) {
        return s.status;
    },
    render: function render() {
        return React.createElement(
            "div",
            { className: "status" },
            React.createElement(MessageLoading, { loading: this.state.f.get('statusLoading'),
                text: this.state.f.get('statusText') }),
            React.createElement(StatusLight, { color: this.state.f.get('connectionColor'),
                text: this.state.f.get('connectionText') })
        );
    }
});

var MessageLoading = React.createClass({
    displayName: "MessageLoading",

    render: function render() {
        var display = "none";
        if (this.props.loading) {
            display = "block";
        }

        var style = {
            "display": display
        };

        return React.createElement(
            "div",
            null,
            React.createElement("div", { style: style, className: "pull-left indicator" }),
            React.createElement(
                "div",
                { className: "pull-left" },
                React.createElement(
                    "p",
                    null,
                    this.props.text
                )
            )
        );
    }
});

var StatusLight = React.createClass({
    displayName: "StatusLight",

    render: function render() {
        var lightClass = "light pull-right ";
        lightClass += this.props.color;

        return React.createElement(
            "div",
            null,
            React.createElement(
                "div",
                { className: "pull-right" },
                React.createElement(
                    "p",
                    null,
                    this.props.text
                )
            ),
            React.createElement("div", { className: lightClass })
        );
    }
});

module.exports = Status;

},{"../components":2}],36:[function(require,module,exports){
"use strict";

var S = require("../store");
var Routes = require("../routes");
var Components = require("../components");
var API = require("../services/api");

var Images = require("../images");

var Toolbar = Components.createStateful({
    stateName: "f",
    filterState: function filterState(s) {
        return {
            identity: s.identity,
            profile: s.profile
        };
    },
    componentWillMount: function componentWillMount() {
        // get current profile and identity
        this.state.f.profile.getCurrent(S);
        this.state.f.identity.load(S);
    },
    toggleNewsfeed: function toggleNewsfeed() {
        S.dispatch(S.actions.views.newsfeed);
    },
    toggleMenu: function toggleMenu() {
        S.dispatch(S.actions.views.menu);
    },
    goHome: function goHome() {
        S.goto(Routes.home);
    },
    goProfile: function goProfile() {
        S.goto(Routes.profile, {});
    },
    goBack: function goBack() {
        S.dispatch(S.actions.url.back);
    },
    goForward: function goForward() {
        S.dispatch(S.actions.url.forward);
    },
    render: function render() {
        var avatar;
        if (this.state.f.profile.store.image !== undefined) {
            var url = API.endpoints.data(this.state.f.profile.store.image);
            avatar = React.createElement("div", { className: "avatar",
                style: Images.background(url) });
        }

        var nickname;
        if (this.state.f.identity.current) {
            nickname = this.state.f.identity.current.Nickname;
        }

        return React.createElement(
            "div",
            { className: "toolbar" },
            React.createElement(
                "div",
                { className: "toolbar-left" },
                React.createElement(
                    "div",
                    { onClick: this.toggleMenu, className: "toolbar-icon" },
                    React.createElement("i", { className: "fa fa-fw fa-bars" })
                ),
                React.createElement(
                    "div",
                    { onClick: this.goHome, className: "toolbar-icon" },
                    React.createElement("i", { className: "fa fa-fw fa-home" })
                ),
                React.createElement(
                    "div",
                    { onClick: this.goBack, className: "toolbar-icon" },
                    React.createElement("i", { className: "fa fa-fw fa-chevron-left" })
                ),
                React.createElement(
                    "div",
                    { onClick: this.goForward, className: "toolbar-icon" },
                    React.createElement("i", { className: "fa fa-fw fa-chevron-right" })
                )
            ),
            React.createElement(
                "div",
                { className: "toolbar-right" },
                React.createElement(
                    "div",
                    { className: "toolbar-icon" },
                    React.createElement("i", { className: "fa fa-fw fa-comments text-danger" })
                ),
                React.createElement(
                    "div",
                    { className: "toolbar-icon" },
                    React.createElement("i", { onClick: this.toggleNewsfeed, className: "fa fa-fw fa-rss" })
                ),
                React.createElement(
                    "div",
                    { onClick: this.goProfile, className: "toolbar-item" },
                    avatar,
                    nickname
                )
            )
        );
    }
});

module.exports = Toolbar;

},{"../components":2,"../images":3,"../routes":15,"../services/api":16,"../store":24}]},{},[1]);
