viewData = (function(window, document, undefined) {
  var data = retrieve('/api/v1/view/transactions', 'error fetching view data');

  return {
    txns: undefined,

    initialize() {
      document.title += ' - Market Transactions'
      data.then(data => {
        this.txns = data.transactions.map(t => setOrderFields(t, data.types, data.stations));
      });
    },
  }
})(window, document, undefined);
