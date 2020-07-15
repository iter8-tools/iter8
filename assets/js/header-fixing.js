(function ($) {
  'use strict';

  // Bootstrap Fixed Header
  $(function() {
    if (('.js-navbar-scroll')) {
      var $this = $('.js-navbar-scroll'),
          onScrollClasses = ($this.data('onscroll-classes')) ? $this.data('onscroll-classes') : 'navbar-bg-onscroll',
          offserValue =  ($this.data('offset-value')) ? $this.data('offset-value') : '150';

      // Check to see if there is a bakcground class on loading
      if ($this.offset().top > offserValue) {
        $this.addClass(onScrollClasses);
      }

      // Check to add a background class on scrolling
      $(window).on('scroll', function() {
        var navbarOffset = $this.offset().top > offserValue;

        if(navbarOffset) {
          $this.addClass(onScrollClasses);
        }
        else {
          $this.removeClass(onScrollClasses);
        }
      });
    }
  });
})(jQuery);