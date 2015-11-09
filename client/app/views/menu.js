var S = require('../store');
var Routes = require('../routes');
var API = require('../services/api');

var Components = require('../components');
var Images = require('../images');

var App = React.createClass({
    openApp: function() {
        console.log("Opening", this.props.plugin.id, this.props.plugin.name);
        
        S.dispatch(S.actions.views.menu);
        
        S.goto(Routes.appView, this.props.plugin);
    },
    render: function() {
        var image = this.props.plugin.image;
        if (image == undefined) {
            image = "/img/icon.png";
        }
        
        var style = {
            "backgroundImage": "url(\"" + encodeURI(image) +  "\")"
        }
        
        return (
            <div onClick={ this.openApp } className="app">
                <div className="app-icon" style={ style }></div>
                <p>{ this.props.plugin.name }</p>
            </div>
        );
    }
});

var Profile = React.createClass({
    renderData: function() {
        var current = this.props.identity.current;
        var aliases = this.props.identity.loadedAliases;

        var address = "";
        if (aliases.length > 0) {
            address = aliases[0].Username + "@" + aliases[0].Location;
        }

        var avatar;
        if (this.props.profile.image) {
            var url = API.endpoints.data(this.props.profile.image);
            avatar = (
                <div className="avatar"
                     style={ Images.background(url) }/>
            );
        }
        
        return (
            <div>
                { avatar }
                <h1>{ current.Nickname }</h1>
                <h2>{ address }</h2>
            </div>
        );
    },
    render: function() {
        var data = (
            <div/>
        );
        if (this.props.identity.hasData && this.props.profile !== undefined) {
            data = this.renderData();
        }
        
        return (          
            <div className="menu-profile">
                <h4>Signed In As</h4>
                { data }
            </div>
        )
    }
})

var Menu = Components.createStateful({
    stateName: "f",
    filterState: function(s) {
        return {
            plugins: s.plugins,
            identity: s.identity,
            profile: s.profile,
        };
    },
    gotoSettings: function() {
        S.dispatch(S.actions.views.menu);
        S.goto(Routes.settings);
    },
    gotoMarket: function() {
        S.dispatch(S.actions.views.menu);
        S.goto(Routes.market);
    },
    closeMenu: function() {
        S.dispatch(S.actions.views.menu);
    },
    componentWillMount: function() {
        // We kick off a loading of the plugins.
        // We will get updated (through state) if this changes without us.
        this.state.f.plugins.loadAll(S);
        this.state.f.identity.load(S);
        this.state.f.profile.getCurrent(S);
    },
    render: function() {
        var allApps = this.state.f.plugins.store.filter(function(p) {
            return !p.hideSidebar;
        }).map(function(p) {
            return (
                <App plugin={ p }/>
            );
        });
        
        return (
            <div className="menu">
                <div onClick={ this.closeMenu } className="close-menu">
                    <i className="fa fa-fw fa-times"></i>
                    Close Menu
                </div>

                <Profile identity={ this.state.f.identity }
                         profile={ this.state.f.profile.store }/>

                <div className="apps">
                    <h4>Apps</h4>

                    <div className="app-container">
                        { allApps }
                    </div>
                </div>

                <div className="statics">
                    <div onClick={ this.gotoMarket } className="static"><p>
                        <i className="fa fa-fw fa-shopping-cart"></i>
                        Marketplace
                    </p></div>
                    <div onClick={ this.gotoSettings } className="static"><p>
                        <i className="fa fa-fw fa-gears"></i>
                        Settings
                    </p></div>
                </div>
            </div>
        );
    },
});

module.exports = Menu
