var Components = require('../../components');

var IdentityOptions = React.createClass({
    render: function() {
        return (
            <div className="identity-options">
                <hr/>
                <div className="inner">
                    <h5>Options</h5>
                </div>
            </div>
        )
    }
});

var Identity = Components.createStateful({
    stateName: "f",
    filterState: function(s) {
        return {
            profile: s.profile,
            identity: s.identity,
        }
    },
    componentWillMount: function() {
        this.state.f.profile.getCurrent(S);
        this.state.f.identity.load(S);
    },
    getInitialState: function() {
        return {
            open: false,
        }
    },
    toggleSettings: function() {
        this.setState({
            open: !this.state.open,
        });
    },
    render: function() {
        console.log(this.state);
        
        var identityClass = "identity";
        if (this.props.active === true) {
            identityClass += " active";
        } else {
            identityClass += " inactive";
        }

        var open = undefined;
        if (this.state.open) {
            open = (
                <IdentityOptions/>
            );
        }
        
        return (
            <div className={ identityClass }>
                <div onClick={ this.toggleSettings } className="identity-settings">
                    <i className="fa fa-cog"></i>
                </div>
                
                <div className="panel">
                    <div className="avatar" style={ getBackground("hunter") }>
                    </div>

                    <div className="go-icon">
                        <i className="fa fa-chevron-right"></i>
                    </div>

                    <div className="content">
                        <h4>Hunter Leath</h4>
                        <p className="text-muted">
                            hleath@airdispat.ch
                        </p>
                    </div>

                    { open }
                </div>
            </div>
        )
    }
})

var Settings = React.createClass({
    render: function() {
        var identities = [
            <Identity/>,
            <Identity/>,
            <Identity/>,
        ]
        
        return (
            <div className="settings">
                <h1>My Identities</h1>

                <h3>Current User</h3>

                <Identity active={true}/>

                <h3>Switch User <small>or <a href="">Create New User</a></small></h3>

                { identities }
            </div>
        );
    }
});

module.exports = Settings;
