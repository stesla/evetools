viewData = (function(window, document, undefined) {
  function descendingByOrderID(a, b) {
    return b.order_id < a.order_id ? -1 : 1;
  }

  let path = window.location.pathname;
  let isCurrent = path.startsWith('/orders');
  var data
  if (isCurrent)
    data = retrieve('/api/v1/view/marketOrders', 'error fetching orders');
  else
    data = retrieve('/api/v1/view/marketOrders?days=30', 'error fetching history');
 
  return {
    initialize() {
      document.title += isCurrent ? ' - Market Orders' : ' - Market Order History';

      data.then(data => {
        this.orders = {
          buy: data.buy.map(o => setOrderFields(o, data.types, data.stations)),
          sell: data.sell.map(o => setOrderFields(o, data.types, data.stations)),
        };
      });
    },

    get sections() {
      return [
        {
          name: "Sell Orders",
          orders: this.orders && this.orders.sell.sort(isCurrent ? byTypeName : descendingByOrderID),
        },
        {
         name: "Buy Orders",
          orders: this.orders && this.orders.buy.sort(isCurrent ? byTypeName : descendingByOrderID),
        },
      ]
    }
  }
})(window, document, undefined);
