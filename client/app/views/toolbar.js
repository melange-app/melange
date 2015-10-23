var S = require("../store");
var Routes = require("../routes");
var Components = require("../components");
var API = require("../services/api");

var Images = require("../images");

var Toolbar = Components.createStateful({
    stateName: "f",
    filterState: function(s) {
        return {
            identity: s.identity,
            profile: s.profile
        }
    },
    componentWillMount: function() {
        // get current profile and identity
        this.state.f.profile.getCurrent(S);
        this.state.f.identity.load(S);
    },
    toggleNewsfeed: function() {
        S.dispatch(S.actions.views.newsfeed);
    },
    toggleMenu: function() {
        S.dispatch(S.actions.views.menu);
    },
    goHome: function() {
        S.goto(Routes.home);
    },
    goProfile: function() {
        S.goto(Routes.profile, {});
    },
    goBack: function() {
        S.dispatch(S.actions.url.back);
    },
    goForward: function() {
        S.dispatch(S.actions.url.forward);
    },
    render: function() {
        var avatar;
        if (this.state.f.profile.store.image !== undefined) {    
            var url = API.endpoints.data(this.state.f.profile.store.image);
            avatar = (
                <div className="avatar"
                     style={ Images.background(url) }></div>
            );
        }

        var nickname;
        if (this.state.f.identity.current) {
            nickname = this.state.f.identity.current.Nickname;
        }
        
        return (
            <div className="toolbar">
                <div className="toolbar-left">
                    <div onClick={ this.toggleMenu } className="toolbar-icon">
                        <i className="fa fa-fw fa-bars"></i>
                    </div>
                    <div onClick={ this.goHome } className="toolbar-icon">
                        <i className="fa fa-fw fa-home"></i>
                    </div>
                    <div onClick={ this.goBack } className="toolbar-icon">
                        <i className="fa fa-fw fa-chevron-left"></i>
                    </div>
                    <div onClick={ this.goForward } className="toolbar-icon">
                        <i className="fa fa-fw fa-chevron-right"></i>
                    </div>
                </div>
                <div className="toolbar-right">
                    <div className="toolbar-icon">
                        <i className="fa fa-fw fa-comments text-danger"></i>
                    </div>
                    <div className="toolbar-icon">
                        <i onClick={ this.toggleNewsfeed} className="fa fa-fw fa-rss"></i>
                    </div>
                    <div onClick={ this.goProfile } className="toolbar-item">
                        { avatar }
                        { nickname }
                    </div>
                </div>
            </div>
        );
    }
});

module.exports = Toolbar;
