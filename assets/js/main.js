(function ($) {
  'use strict';

  // Activate Tooltips & Popovers
  $(function () {
    $('[data-toggle="tooltip"]').tooltip();
    $('[data-toggle="popover"]').popover();

    // Dismiss Popovers on next click
    $('.popover-dismiss').popover({
      trigger: 'focus'
    })
  });

  $(document).on('ready', function () {
    // Go to Top
    var go2TopShowHide = (function () {
      var $this = $('.js-go-to');

      $this.on("click", function(event) {
        event.preventDefault();
        $("html, body").animate({scrollTop: 0}, 600);
      });

      var go2TopOperation = function() {
        var CurrentWindowPosition = $(window).scrollTop();

        if (CurrentWindowPosition > 400) {
          $this.addClass("show");
        } else {
          $this.removeClass("show");
        }
      };

      go2TopOperation();

      $(window).scroll(function() {
        go2TopOperation();
      });
    }());
  });

  $(window).on('load', function () {

    // Page Nav
    var onePageScrolling = (function () {
      $('.js-scroll-nav a').on('click', function(event) {
        event.preventDefault();
        if ( $('.duik-header').length ) {
          $('html, body').animate( {scrollTop:( $('#' + this.href.split('#')[1]).offset().top - ( $('.duik-header .navbar').height() ) - 30 )}, 600 );
        } else {
          $('html, body').animate( {scrollTop:( $('#' + this.href.split('#')[1]).offset().top - 30 )}, 600 );
        }
      });
    }());

    var oneAnchorScrolling = (function () {
      $('.js-anchor-link').on('click', function(event) {
        event.preventDefault();
        if ( $('.duik-header').length ) {
          $('html, body').animate( {scrollTop:( $('#' + this.href.split('#')[1]).offset().top - ( $('.duik-header .navbar').height() ) - 30 )}, 600 );
        } else {
          $('html, body').animate( {scrollTop:( $('#' + this.href.split('#')[1]).offset().top - 30 )}, 600 );
        }
      });
    }());
  });
})(jQuery);