// Used in assets/documentation/introduction/tutorial-deployments.html to control tabs

(function ($) {
  'use strict';

  $(document).on('ready', function () {
    // Change tabs for links that go to those tabs
    $("a[href='#nav-part1']").click(function () {
      $('#nav-tab a[href="#nav-part1"]').tab('show');
    })

    $("a[href='#nav-part2']").click(function () {
      $('#nav-tab a[href="#nav-part2"]').tab('show');
    })

    $("a[href='#nav-part3']").click(function () {
      $('#nav-tab a[href="#nav-part3"]').tab('show');
    })

    $("a[href='#nav-part4']").click(function () {
      $('#nav-tab a[href="#nav-part4"]').tab('show');
    })

    $("a[href='#nav-part5']").click(function () {
      $('#nav-tab a[href="#nav-part5"]').tab('show');
    })
  });
})(jQuery);