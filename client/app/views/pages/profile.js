var S = require('../../store');
var Components = require('../../components');
var Images = require('../../images');
var Message = require('../message');

var Backend = require('../../services/backend');
var API = Backend.api;

var Profile = Components.createStateful({
    stateName: "f",
    filterState: function(s) {
        return {
            profile: s.profile,
            messages: s.messages,
        };
    },
    componentWillMount: function() {
        if (this.props.data.has('id')) {
            // Download the other user's messages and profile.
            Backend.messages.get(this.props.data.get('id'), "profile");
        } else {
            this.state.f.profile.getCurrent(S);
        }
    },
    renderBackdrop: function(profile) {
        var backdrop = {
            "backgroundColor": "#1A8BB0",
        };
        
        if (profile.backdrop) {
            var url = API.endpoints.data(profile.backdrop);
            backdrop = Images.background(url);
        }
        
        return (
            <div className="backdrop"
                 style={ backdrop }/>
        );
    },
    renderAvatar: function(profile) {
        var avatar;
        if (profile.image) {
            var url = API.endpoints.data(profile.image);
            avatar = (
                <div className="avatar"
                     style={ Images.background(url) }/>
            );
        }

        return avatar;
    },
    render: function() {
        var profile = this.state.f.profile.store;
        profile.messages = this.state.f.messages.store.filter(function(val) {
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
                    image: "airdispat.ch/profile/avatar",
                };
                
                profile = {
                    name: components.get(names.name).get('string'),
                    description: components.get(names.description).get('string'),
                    image: components.get(names.image).get('string'),
                    messages: this.state.f.messages.store.filter(function(val) {
                        return val.getIn(['from', 'alias']) == alias && val.get('public');
                    }),
                };
            } else {
                profile = {
                    name: "Loading...",
                }
            }
        }

        var description;
        if (profile.description) {
            description = profile.description;
        }

        var messages = profile.messages.map(function(val) {
            return (
                <Message message={ val }/>
            )
        });

        return (
            <div className="profile">
                { this.renderBackdrop(profile) }
                { this.renderAvatar(profile) }
                
                <div className="info">
                    <h1>{ profile.name }</h1>
                    <p className="muted">
                        <a href="">Edit My Profile</a>
                    </p>
                    <p>
                        { description }
                    </p>
                </div>

                <div className="profile-feed">
                    { messages }
                </div>
            </div>
        );
    }
});

module.exports = Profile;
