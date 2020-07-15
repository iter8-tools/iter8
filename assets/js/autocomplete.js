(function ($) {
  'use strict';

  // custom dropdown for autocomplete
  $.widget('custom.localcatcomplete', $.ui.autocomplete, {
    _create: function () {
      this._super();
      this.widget().menu('option', 'items', '> :not(.ui-autocomplete-category)');
    },
    _renderItem: function (ul, item) {
      var label = item.label ? '<span class="duik-search__label">' + item.label + '</span>' : '',
        innerText = label + '<span class="duik-search__category">' + item.category + '</span>';

      if (item.url) {
        return $('<li><a href="' + window.location.protocol + '//' + window.location.host + '/' + window.location.pathname.split('/')[1] + '/' + item.url + '">' + innerText + '</a></li>')
          .appendTo(ul);
      } else {
        return $('<li>' + item.label + '</li>')
          .appendTo(ul);
      }
    }
  });

  $(document).on('ready', function () {
    // Autocomplete
    $('.js-search').each(function (i, el) {
      var $this = $(el),
        dataUrl = $this.data('url');

      $.getJSON(dataUrl, function (data) {
        $this.localcatcomplete({
          appendTo: $this.parent(),
          delay: 0,
          source: data,
          select: function (event, ui) {
            window.location = window.location.protocol + '//' + window.location.host + '/' + window.location.pathname.split('/')[1] + '/' + ui.item.url;
          }
        });
      });
    });
  });
})(jQuery);