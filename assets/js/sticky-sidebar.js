if (('.js-sticky-sidebar')) {
  var stickySidebar = new StickySidebar('.js-sticky-sidebar', {
      topSpacing: 30,
      bottomSpacing: 30,
  });

  $('.js-sticky-sidebar a[data-toggle="tab"]').on('shown.bs.tab', function (e) {
    stickySidebar.updateSticky();
  })
}