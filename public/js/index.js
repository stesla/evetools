viewData = (function(window, document, undefined) {

  return {
    data: undefined,
    favorites: [],
    walletBalance: 0,
    buyTotal: 0,
    sellTotal: 0,

    initialize() {
      document.title += " - Dashboard"
      evetools.currentUser
      .then(user => {
        this.user = user;
        this.walletBalance = user.wallet_balance;
      })
      .then(() => {
        return retrieve('/api/v1/user/orders', 'error fetching market orders');
      })
      .then(orders => {
        this.buyTotal = orders.buy.reduce((a, o) => a + o.escrow, 0);
        this.sellTotal = orders.sell.reduce((a, x) => a + x.volume_remain * x.price, 0);
      });

      evetools.sdeTypes().then(types => {
        this.favorites = this.user.favorites.map(id => {
          let type = types[""+id];
          type.favorite = true;
          return type;
        }).sort(byName);
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
