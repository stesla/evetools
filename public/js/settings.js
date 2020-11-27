viewData = (function(window, document, undefined) {
  function stationList(slot, name, id) {
    return {
      editing: false,
      listOpen: false,
      station: { name: name, id: id },
      stationName: "",
      stationList: [],

      beginEdit(event, nextTick) {
        this.editing = true
        let input = event.target.parentElement.parentElement.parentElement.querySelector('input');
        nextTick(() => {
          input.focus();
        });
      },

      fetch() {
        if (this.stationName.length < 3) {
          this.stationList = [];
          return;
        }
        const params = new URLSearchParams();
        params.set("q", this.stationName);
        retrieve('/api/v1/stations?' + params.toString()).then(stationList => {
          this.stationList = stationList.sort(byName);
        });
      },

      get loaded() {
        return loaded;
      },

      save() {
        if (this.stationName === "") {
          this.stationList = [];
          this.editing = false;
          return;
        }
        station = this.stationList.find(s => s.name === this.stationName);
        return retrieve('/api/v1/user/'+slot, 'error saving station', {
          raw: true,
          method: 'PUT',
          body: JSON.stringify(station),
        })
        .then(() => {
          this.station = station;
          this.stationName = "";
          this.stationList = [];
          this.editing = false;
        });
      },

      select(event, nextTick) {
        this.stationName=event.target.value;
        this.listOpen=false;
        let button = event.target.parentElement.parentElement.querySelector('button');
        nextTick(() => { button.focus(); });
      },
    };
  }
  return stationList
})(window, document, undefined);
