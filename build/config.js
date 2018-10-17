import path from 'path'

/**
 * project base path
 */
const basePath = path.join(__dirname, '../');

export default {

    /**
     * Absolute paths to known dirs, e.g., to src dir
     */
    paths: {
        app: path.join(basePath, 'src/app'),
        assets: path.join(basePath, 'src/app/assets'),
        backendSrc: path.join(basePath, 'src/app/backend'),
        base: basePath,
        bowerComponents: path.join(basePath, 'bower_components'),
        build: path.join(basePath, 'build'),
        dist: path.join(basePath, 'dist'),
        externs: path.join(basePath, 'src/app/externs'),
        frontendSrc: path.join(basePath, 'src/app/frontend'),
        frontendTest: path.join(basePath, 'src/test/frontend'),
        karmaConf: path.join(basePath, 'build/karma.conf.js'),
        nodeModules: path.join(basePath, 'node_modules'),
        partials: path.join(basePath, '.tmp/partials'),
        prodTmp: path.join(basePath, '.tmp/prod'),
        serve: path.join(basePath, '.tmp/serve'),
        src: path.join(basePath, 'src'),
        tmp: path.join(basePath, '.tmp')
    },

    frontend: {
        // The name of Angular module, i.e., the module that bootstrap the application
        rootModuleName: 'k8sconsole'
    }
};
