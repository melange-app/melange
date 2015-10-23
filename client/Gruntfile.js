module.exports = function(grunt) {
    grunt.initConfig({
        browserify: {
            options: {
                transform: [ require('babelify') ]
            },
            app: {
                src: 'app/app.js',
                dest: 'js/melange.js'
            },
        },
        sass: {
            app: {
                files: {
                    'css/melange.css': ['style/melange.scss']
                }
            }
        },
        watch: {
            options: {
                spawn: false,
                dot: false,
            },
            css: {
                files: [
                    'style/**/*.scss',
                    '!style/**/.*'
                ],
                tasks: ['sass']
            },
            js: {
                files: [
                    'app/**/*.js',
                    '!app/**/.*'
                ],
                tasks: ['compileJS']
            }
        },
    })

    grunt.loadNpmTasks('grunt-contrib-sass');
    grunt.loadNpmTasks('grunt-contrib-watch');
    grunt.loadNpmTasks('grunt-browserify');

    grunt.registerTask('compileJS', ['browserify']);
    grunt.registerTask('compile', ['compileJS', 'sass']);
}
