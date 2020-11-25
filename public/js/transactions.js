viewData = (function(window, document, undefined) {
  var data = retrieve('/api/v1/view/transactions', 'error fetching view data');

  const leadingZero = n => n > 9 ? n : '0' + n;

  function formatDate(str) {
    let date = new Date(str);
    return date.getFullYear() + '-' +
        leadingZero(date.getMonth() + 1) + '-' +
        leadingZero(date.getDate()) + ' ' +
        leadingZero(date.getHours()) + ':' +
        leadingZero(date.getMinutes()) + ':' +
        leadingZero(date.getSeconds());
  }

  return {
    txns: undefined,

    initialize() {
      document.title += ' - Market Transactions'
      data.then(data => {
        this.txns = data.transactions.map(t => {
          t.date = formatDate(t.date);
          t.station_name = data.stations[''+t.location_id].name;
          t.type = data.types[''+t.type_id];
          return t
        });
      });
    },
  }
})(window, document, undefined);
