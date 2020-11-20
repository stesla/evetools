viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/types/(.*)");
  let match = window.location.pathname.match(typeRE);
  let typeID = match[1];

  var data = retrieve('/api/v1/view/typeDetails/'+typeID, 'error fetching view data');

  return {
    type: undefined,
    typeID: typeID,

    toggleFavorite() {
      setFavorite(typeID, !this.favorite)
      .then(obj => {
        this.favorite = obj.favorite;
      });
    },

    initialize() {
      const observer = new MutationObserver(() => {
        let div = document.getElementById("chart");
        if (div) {
          observer.disconnect();
          renderChart(this.info.history, 400, div.clientWidth);
        }
      });

      observer.observe(document.querySelector('main'), { childList: true, subtree: true });

      data.then(data => {
        document.title += ' - ' + data.type.name;
        this.type = data.type;
        this.station = data.station;
        this.system = data.system;
        this.favorite = data.favorite;
        data.marketInfo.history = data.marketInfo.history.map(d => {
          return {
            date: Date.parse(d.date),
            average: +d.average,
          }
        });
        this.info = data.marketInfo;
        this.group = data.group;
        this.parentGroups = data.parent_groups.reverse();
      });
    }
  }
})(window, document);

