viewData = (function(window, document, undefined) {
  var types = retrieve('/data/types.json', 'error fetching sde types');
  var stations = retrieve('/data/stations.json', 'error fetching sde stations');

  return {
    txns: undefined,

    initialize() {
      document.title += ' - Market Transactions'
      types.then(types => {
        this.types = types;
        return stations;
      })
      .then(stations => {
        this.stations = stations;
        return  retrieve('/api/v1/user/transactions', 'error fetching wallet transactions')
      })
      .then(txns => {
        this.txns = txns.map(t => setOrderFields(t, this.types, this.stations));
      });
    },
  }
})(window, document, undefined);
