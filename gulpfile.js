//
// Gulpfile
//
var gulp                   = require('gulp'),
    sass                   = require('gulp-sass'),
    changed                = require('gulp-changed'),
    autoprefixer           = require('gulp-autoprefixer'),
    rename                 = require('gulp-rename'),
    del                    = require('del'),
    concat                 = require('gulp-concat'),
    cssnano                = require('gulp-cssnano'),
    uglify                 = require('gulp-uglifyjs'),
    cache                  = require('gulp-cache'),
    imagemin               = require('gulp-imagemin'),
    imageminJpegRecompress = require('imagemin-jpeg-recompress'),
    pngquant               = require('imagemin-pngquant'),
    browserSync            = require('browser-sync').create();



//
// Gulp plumber error handler - displays if any error occurs during the process on your command
//
function errorLog(error) {
  console.error.bind(error);
  this.emit('end');
}



//
// SASS - Compile SASS files into CSS
//
gulp.task('sass', function () {
 // Theme
 gulp.src('./assets/include/scss/**/*.scss')
  .pipe(changed('./assets/css/'))
  .pipe(sass({ outputStyle: 'expanded' }))
  .on('error', sass.logError)
  .pipe(autoprefixer([
      "last 1 major version",
      ">= 1%",
      "Chrome >= 45",
      "Firefox >= 38",
      "Edge >= 12",
      "Explorer >= 10",
      "iOS >= 9",
      "Safari >= 9",
      "Android >= 4.4",
      "Opera >= 30"], { cascade: true }))
  .pipe(gulp.dest('./assets/css/'))
  .pipe(browserSync.stream());
});



//
// BrowserSync (live reload) - keeps multiple browsers & devices in sync when building websites
//
//
gulp.task('serve', function() {
  browserSync.init({
    files: "./*.html",
    startPath: "./index.html",
    server: {
      baseDir: "./",
      routes: {},
      middleware: function (req, res, next) {
        if (/\.json|\.txt|\.html/.test(req.url) && req.method.toUpperCase() == 'POST') {
          console.log('[POST => GET] : ' + req.url);
          req.method = 'GET';
        }
        next();
      }
    }
  })
});



//
// Gulp Watch and Tasks
//
//
gulp.task('watch', function() {
  gulp.watch('./assets/include/scss/**/*.scss', ['sass']);
  gulp.watch('./html/**/*.html').on('change', browserSync.reload);
  gulp.watch('./starter/**/*.html').on('change', browserSync.reload);
  gulp.watch('./documentation/**/*.html').on('change', browserSync.reload);
});

// Gulp Tasks
gulp.task('default', ['watch', 'sass', 'serve'])



//
// CSS minifier - merges and minifies the below given list of Front libraries into one theme.min.css
//
gulp.task('minCSS', function() {
  return gulp.src([
    './assets/css/theme.css',
  ])
  .pipe(cssnano())
  .pipe(rename({suffix: '.min'}))
  .pipe(gulp.dest('./dist/assets/css/'));
});



//
// JavaSript minifier - merges and minifies the below given list of Front libraries into one theme.min.js
//
gulp.task('minJS', function() {
  return gulp.src([
    './assets/js/main.js',
    './assets/js/autocomplete.js',
    './assets/js/custom-scrollbar.js',
    './assets/js/sticky-sidebar.js',
    './assets/js/header-fixing.js',
    './assets/js/theme-custom.js'
  ])
  .pipe(concat('theme.min.js'))
  .pipe(uglify())
  .pipe(gulp.dest('./dist/assets/js/'));
});


//
// Image minifier - compresses images automatically
//

gulp.task('minIMG', function() {
  return gulp.src('./assets/img-temp/**/*')
    .pipe(cache(imagemin([
      imagemin.gifsicle({interlaced: true}),
      imagemin.jpegtran({progressive: true}),
      imageminJpegRecompress({
        loops: 5,
        min: 65,
        max: 70,
        quality:'medium'
      }),
      imagemin.svgo(),
      imagemin.optipng({optimizationLevel: 3}),
      pngquant({quality: '65-70', speed: 5})
    ],{
      verbose: true
    })))
    .pipe(gulp.dest('./dist/assets/img-temp/'));
});


gulp.task('dist', ['minCSS', 'minJS', 'minIMG']);