viewData = (function(window, document, undefined) {
  let typeRE = new RegExp("/types/(.*)");
  let match = window.location.pathname.match(typeRE);
  let typeID = match[1];

  var data = retrieve('/api/v1/view/typeDetails/'+typeID, 'error fetching view data');

  function chartPoint(d) {
    return {
      date: Date.parse(d.date),
      average: +d.average,
    }
  }

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
        let div = document.getElementById("chartA");
        if (div) {
          observer.disconnect();
          renderChart("#chartA", this.infoA.history, 400, div.clientWidth);
        }
      });

      observer.observe(document.querySelector('main'), { childList: true, subtree: true });

      data.then(data => {
        document.title += ' - ' + data.type.name;
        this.type = data.type;
        this.favorite = data.favorite;
        this.group = data.group;
        this.parentGroups = data.parent_groups.reverse();

        this.infoA = data.infoA;
        this.infoA.history = data.infoA.history.map(d => chartPoint(d));

        this.infoB = data.infoB;
        this.infoB.history = data.infoB.history.map(d => chartPoint(d));
      });
    }
  }
})(window, document);

