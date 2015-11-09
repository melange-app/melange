var routes = {
    home: {
        name: "home",
        page: require('./views/pages/home'),
    },
    profile: {
        name: "profile",
        page: require('./views/pages/profile'),
    },
    setup: {
        name: "setup",
        page: require('./views/pages/setup'),
    },
    settings: {
        name: "settings",
        page: require('./views/pages/settings'),
    },
    market: {
        name: "market",
        page: require('./views/pages/market'),
    },
    appView: {
        name: "appView",
        page: require('./views/plugin').Page,
    },
};

routes.notFound = {
    name: "notFound",
    page: require('./views/pages/notFound'),
}

routes.default = routes.home;

module.exports = routes;
