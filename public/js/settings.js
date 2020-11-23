viewData = (function(window, document, undefined) {
  var data = retrieve('/api/v1/view/settings', 'error fetching data view');

  return {
    characters: {},
    editingStation: false,
    station: { name: "" },
    stationName: "",
    stations: [],
    loaded: false,

    initialize() {
      document.title += " - Settings";
      data.then(data => {
        this.characters = data.characters;
        this.station = data.station;
        this.loaded = true;
      });
    },

    get characterList() {
      return Object.values(this.characters).sort(byName);
    },

    fetchStations() {
      if (this.stationName.length < 3) {
        return;
      }
      const params = new URLSearchParams();
      params.set("q", this.stationName);
      retrieve('/api/v1/stations?' + params.toString()).then(stations => {
        this.stations = stations.sort(byName);
      });
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

    saveStation() {
      if (this.stationName === "") {
        this.editingStation = false;
        return;
      }
      station = this.stations.find(s => s.name === this.stationName);
      return retrieve('/api/v1/user/station', 'error saving station', {
        raw: true,
        method: 'PUT',
        body: JSON.stringify(station),
      })
      .then(() => {
        this.station = station;
        this.stationName = "";
        this.editingStation = false;
      });
    },
  }
})(window, document, undefined);
