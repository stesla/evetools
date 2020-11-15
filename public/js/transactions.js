viewData = (function(window, document, undefined) {
  return {
    txns: undefined,

    initialize() {
      document.title += ' - Market Transactions'
      evetools.sdeTypes().then(types => {
        this.types = types;
        return evetools.sdeStations();
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
