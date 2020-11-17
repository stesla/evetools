viewData = (function(window, document, undefined) {
  var characters = retrieve('/api/v1/user/characters', 'error fetching characters');
  var currentUser = window.retrieve('/api/v1/user/current', 'error fetching current user');
  var stations = retrieve('/data/stations.json', 'error fetching sde stations');

  return {
    characters: {},
    editingStation: false,
    station: { name: "" },
    stationName: "",
    stations: {},

    initialize() {
      document.title += " - Settings";
      currentUser.then(user => {
        this.user = user;
        return stations;
      })
      .then(stations => {
        this.station = stations[''+this.user.station_id];
      });

      characters.then(list => {
        this.characters = list;
      });
    },

    get characterList() {
      return Object.values(this.characters).sort(byName);
    },

    fetchStations() {
      if (this.stationName.length < 3) {
        return;
      }
      stations.then(stations => {
        this.stations = Object.values(stations).filter(s => {
          return s.name.toLowerCase().includes(this.stationName.toLowerCase());
        }).reduce((m, s) => {
          m[s.name] = s;
          return m;
        }, {});
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
      stations.then(stations => {
        let station = Object.values(stations).find(s => s.name == this.stationName);
        this.station = station;
        return retrieve('/api/v1/user/station', 'error saving station', {
          raw: true,
          method: 'PUT',
          body: JSON.stringify(station),
        });
      })
      .then(() => {
        this.stationName = "";
        this.editingStation = false;
      });
    },

    get stationList() {
      return Object.values(this.stations);
    },
  }
})(window, document, undefined);
