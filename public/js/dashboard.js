viewData = (function(window, document, undefined) {
  var dashboard = retrieve('/api/v1/view/dashboard', 'error fetching view data');

  return {
    favorites: [],
    walletBalance: 0,
    brokerFee: 0,
    buyTotal: 0,
    sellTotal: 0,
    loaded: false,

    initialize() {
      document.title += " - Dashboard"

      dashboard.then(data => {
        this.favorites = data.favorites.map(type => {
          type.favorite = true;
          return type;
        }).sort(byName);
        this.walletBalance = data.wallet_balance;
        this.brokerFee = data.broker_fee;
        this.buyTotal = data.buy_total;
        this.sellTotal = data.sell_total;
        this.loaded = true
      });
    },

    toggleFavorite(type) {
      let val = !type.favorite
      setFavorite(type.id, val)
      .then(() => {
        type.favorite = val
      });
    },
  }
})(window, document, undefined);
