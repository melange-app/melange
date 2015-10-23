var createEndpoint = function(prefix) {
    var melangeSuffix = ".local.getmelange.com:7776";

    if (prefix == "ws") {
        return "ws://api.local.getmelange.com:7776/realtime"
    }
    
    return "http://" + prefix + melangeSuffix;
}

var endpoints = {
    api: createEndpoint("api"),
    common: createEndpoint("common"),
    plugins: function(pluginId) {
        var prefix = encodeURI(pluginId) + ".plugins";
        return createEndpoint(prefix);
    },
    data: function() {
        var prefix = createEndpoint("data");
        
        if (arguments.length == 1) {
            return prefix + "/" + arguments[0];
        } else if (arguments.length == 2) {
            return prefix + "/" + arguments[0] + "/" + arguments[1];
        }

        return prefix;
    },
    ws: createEndpoint("ws"),
}

var getCanonicalAPILocation = function(url) {
    var prefix = endpoints.api;
    
    if (url instanceof String) {
        return prefix + url;
    } else if (url instanceof Array) {
        for (var i in url) {
            prefix += "/" + url[i];
        }

        return prefix;
    } else if (url instanceof Immutable.List) {
        return url.reduce(function(acc, val) {
            return acc + "/" + val;
        }, prefix);
    }

    return undefined;
}

var call = function(method, url, data) {
    var deferred = Q.defer();
    var request = superagent;

    url = getCanonicalAPILocation(url);
    
    if (method == 'get') {
        request = request.get(url).query(data);
    } else if (method == 'post') {
        request = request.post(url).send(data);
    }

    request.type("json").end(function(err, res) {
        if (res.ok) {
            deferred.resolve(res.body);
        } else {
            deferred.reject({
                status: res.status,
                body: res.text,
            })
        }
    });

    return deferred.promise;
}

module.exports = {
    get: function(url, data) {
        return call('get', url, data);
    },
    post: function(url, data) {
        return call('post', url, data);
    },
    endpoints: endpoints,
}
