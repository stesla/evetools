viewData = (function(window, document, undefined) {
  var data = retrieve('/api/v1/view/settings', 'error fetching data view');
  var loaded = false;
  data.then(() => { loaded = true; });

  function settings() {
    return {
      characters: {},

      initialize() {
        document.title += " - Settings";
        data.then(data => {
          this.characters = data.characters;
        });
      },

      get characterList() {
        return Object.values(this.characters).sort(byName);
      },

      get loaded() {
        return loaded;
      },

      makeActiveCharacter(cid) {
        retrieve('/api/v1/user/characters/' + cid + '/activate', 'error activating user', {
          raw: true,
          method: 'POST',
        })
        .then(() => {
          window.location.href = "/";
        });
      },

    }
  }

  function stationList(station) {
    return {
      editing: false,
      listOpen: false,
      station: { name: "" },
      stationName: "",
      stationList: [],

      initialize() {
        data.then(data => {
          this.station = data[station];
        });
      },

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
        return retrieve('/api/v1/user/station', 'error saving station', {
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

  return {
    settings: settings,
    stationList: stationList,
  }
})(window, document, undefined);
