(function ($) {
  'use strict';

  $(document).on('ready', function () {
    // Custom Scroll
    $('.js-scrollbar').mCustomScrollbar({
      theme: 'minimal-dark',
      scrollInertia: 150
    });

    // Scroll to Active
    $('.js-scrollbar').mCustomScrollbar('scrollTo', '.js-scrollbar a.active');
  });
})(jQuery);