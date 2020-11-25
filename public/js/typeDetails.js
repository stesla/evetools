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

  function waitForChart(id, historyfn) {
    const observer = new MutationObserver(() => {
      let div = document.getElementById(id);
      if (div) {
        observer.disconnect();
        let history = historyfn().map(d => chartPoint(d));
        let width = div.parentElement.parentElement.clientWidth;
        renderChart('#' + id, history, 400, width);
      }
    });
    observer.observe(document.querySelector('main'), { childList: true, subtree: true });
  }

  return {
    tab: window.location.hash ? window.location.hash.slice(1,2) : 'a',
    type: undefined,
    typeID: typeID,

    toggleFavorite() {
      setFavorite(typeID, !this.favorite)
      .then(obj => {
        this.favorite = obj.favorite;
      });
    },

    initialize() {
      waitForChart('chartA', () => this.infoA.history);
      waitForChart('chartB', () => this.infoB.history);
      data.then(data => {
        document.title += ' - ' + data.type.name;
        this.type = data.type;
        this.favorite = data.favorite;
        this.group = data.group;
        this.parentGroups = data.parent_groups.reverse();
        this.infoA = data.infoA;
        this.infoB = data.infoB;
      });
    },

    selectTab(tab) {
      this.tab = tab;
      window.location.hash = tab;
    }
  }
})(window, document);

