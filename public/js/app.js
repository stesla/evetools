evetools = (function(document, window, undefined) {
  var result = {};

  function retrieve(url, errmsg, options) {
    return fetch(url, options)
    .then(resp => {
      if(!resp.ok) {
        throw new Error(errmsg);
      }
      return options && options.raw ? resp : resp.json();
    });
  }

  var currentUser = retrieve('/api/v1/user/current', 'error fetching current user');

  var _sdeTypes
  function sdeTypes() {
    if (!_sdeTypes)
      _sdeTypes = retrieve('/data/types.json', 'error fetching sde types');
    return _sdeTypes
  }

  var _sdeMarketGroups
  function sdeMarketGroups() {
    if (!_sdeMarketGroups)
      _sdeMarketGroups = retrieve('/data/marketGroups.json', 'error fetching sde market groups');
    return _sdeMarketGroups;
  }

  var _sdeStations
  function sdeStations() {
    if (!_sdeStations)
      _sdeStations = retrieve('/data/stations.json', 'error fetching sde stations');
    return _sdeStations;
  }

  var _sdeSystems
  function sdeSystems() {
    if (!_sdeSystems)
      _sdeSystems = retrieve('/data/systems.json', 'error fetching sde systems');
    return _sdeSystems;
  }

  result.globalState = function() {
    return {
      avatarMenuOpen: false,
      marketMenuOpen: false,
      loggedIn: false,
      navOpen: false,
      user: {
        character: { name: "", id: 0},
        sationName: "",
      },

      get currentView() {
        if (!this.loggedIn) {
          return 'login';
        }

        let path = window.location.pathname;

        if (path === '/')
          return 'index';

        if (path.startsWith('/browse'))
          return 'browse';

        if (path.startsWith('/groups/'))
          return 'groupDetails';

        if (path.startsWith('/history'))
          return 'history';

        if (path.startsWith('/orders'))
          return 'orders';

        if (path.startsWith('/search'))
          return 'search';

        if (path.startsWith('/transactions'))
          return 'transactions';

        if (path.startsWith('/types/'))
          return 'typeDetails';

        return 'notFound';
      },

      initialize() {
        currentUser.then(user => {
          user.character = user.characters[user.active_character];
          this.user = user;
          this.loggedIn = true;
          return user;
        })
        .catch(() => {})
        .then(() => {
          const url = '/views/'+this.currentView+'.html';
          return retrieve(url, 'error fetching view', { raw: true });
        })
        .then(resp => resp.text())
        .then(html => {
          const parser = new DOMParser();
          const elt = parser.parseFromString(html, 'text/html').querySelector('main');
          const slot = document.querySelector('main');
          slot.parentNode.replaceChild(elt, slot);
        });
      },
    }
  }

  // Views

  function byName(a, b) {
    return a.name < b.name ? -1 : 1;
  }

  function byOrderID(a, b) {
    return a.order_id < b.order_id ? -1 : 1;
  }

  result.browse = function() {
    return {
      data: { root: [] },
      filter: "",

      handleSearch(e) {
        e.preventDefault();
        window.handleSearch(this.filter);
      },

      get groups() {
        return this.data.root.map(id =>
          this.data.groups[''+id]
        ).sort(byName);
      },

      initialize() {
        sdeMarketGroups().then(data => {
          this.data = data
        });
        document.title += " - Find Items"
      }
    }
  }

  result.groupDetails = function() {
    let typeRE = new RegExp("/groups/(.*)");
    let match = window.location.pathname.match(typeRE);

    return {
      group: { name: "", groups: [] },
      groupID: match[1],
      marketGroups: { root: [] },
      types: {},
      filter: "",
      parent: { name: "" },

      get children() {
        if (!this.group || Object.keys(this.types).length == 0)
          return [];

        if (this.group.groups) {
          return this.group.groups.map(id => {
            let g = this.marketGroups.groups[''+id];
            g.isGroup = true;
            return g
          }).sort(byName);
        } else if (this.group.types) {
          return this.group.types.map(id => {
            let t = this.types[''+id];
            t.isType = true;
            return t;
          }).sort(byName);
        }
      },

      initialize() {
        sdeMarketGroups().then(data => {
          this.marketGroups = data;
          this.group = data.groups[''+this.groupID];
          this.parent = data.groups[''+this.group.parent_id];
          document.title += " - " + this.group.name;
        });

        sdeTypes().then(types => {
          this.types = types;
        });
      },
    }
  }

  result.index = function() {
    return {
      data: undefined,
      characters: {},
      editingStation: false,
      favorites: [],
      station: { name: "" },
      stationName: "",
      stations: {},
      walletBalance: 0,
      buyTotal: 0,
      sellTotal: 0,

      initialize() {
        document.title += " - Dashboard"
        currentUser
        .then(user => {
          this.user = user;
          this.characters = user.characters;
          this.walletBalance = user.wallet_balance;
        })
        .then(() => {
          return retrieve('/api/v1/user/orders', 'error fetching market orders');
        })
        .then(orders => {
          this.buyTotal = orders.buy.reduce((a, o) => a + o.escrow, 0);
          this.sellTotal = orders.sell.reduce((a, x) => a + x.volume_remain * x.price, 0);
        });

        sdeStations().then(stations => {
          this.station = stations[''+this.user.station_id];
        });

        sdeTypes().then(types => {
          this.favorites = this.user.favorites.map(id => {
            let type = types[""+id];
            type.favorite = true;
            return type;
          }).sort(byName);
        });
      },

      fetchStations() {
        if (this.stationName.length < 3) {
          return;
        }
        sdeStations().then(stations => {
          this.stations = Object.values(stations).filter(s => {
            return s.name.toLowerCase().includes(this.stationName.toLowerCase());
          }).reduce((m, s) => {
            m[s.name] = s;
            return m;
          }, {});
        });
      },

      get characterList() {
        return Object.values(this.characters).sort(byName);
      },

      get stationList() {
        return Object.values(this.stations);
      },

      toggleFavorite(type) {
        let val = !type.favorite
        setFavorite(type.id, val)
        .then(() => {
          type.favorite = val
        });
      },

      saveStation() {
        if (this.stationName === "") {
          this.editingStation = false;
          return;
        }
        sdeStations().then(stations => {
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
    }
  }

  function byTypeName(a, b) {
    return a.type.name < b.type.name ? -1 : 1;
  }

  function setOrderFields(o, types, stations) {
    let type = types[''+o.type_id]
    let station = stations[''+o.location_id]
    o.type = type;
    o.station_name = station.name;
    return o;
  }

  result.history = function() {
    return {
      initialize() {
        document.title += ' - Market Order History'
        sdeTypes().then(types => {
          this.types = types;
          return sdeStations();
        })
        .then(stations => {
          this.stations = stations;
          return retrieve('/api/v1/user/history?days=30', 'error fetching history');
        })
        .then(data => {
          this.orders = {
            buy: data.buy.map(o => setOrderFields(o, this.types, this.stations)),
            sell: data.sell.map(o => setOrderFields(o, this.types, this.stations)),
          };
        });
      },

      get sections() {
        return [
          {
            name: "Sell Orders",
            orders: this.orders && this.orders.sell.sort(byOrderID).reverse()
          },
          {
            name: "Buy Orders",
            orders: this.orders && this.orders.buy.sort(byOrderID).reverse()
          },
        ];
      }
    }
  }

  result.orders = function() {
    return {
      initialize() {
        document.title += ' - Market Orders'
        sdeTypes().then(types => {
          this.types = types;
          return sdeStations();
        })
        .then(stations => {
          this.stations = stations;
          return retrieve('/api/v1/user/orders', 'error fetching orders');
        })
        .then(data => {
          this.orders = {
            buy: data.buy.map(o => setOrderFields(o, this.types, this.stations)),
            sell: data.sell.map(o => setOrderFields(o, this.types, this.stations)),
          };
        });
      },

      get sections() {
         return [
          {
            name: "Sell Orders",
            orders: this.orders && this.orders.sell.sort(byTypeName),
          },
          {
           name: "Buy Orders",
            orders: this.orders && this.orders.buy.sort(byTypeName),
          },
        ]
      }
    }
  }

  result.search = function() {
    const urlParams = new URLSearchParams(window.location.search);
    return {
      filter: urlParams.get('q'),
      marketTypes: [],

      fetchData() {
        sdeTypes().then(types => {
          ids = Object.values(types).filter(t => {
            let filter = this.filter.toLowerCase();
            return t.name.toLowerCase().includes(filter);
          }).map(t => t.id);
          this.marketTypes = ids.map(id => types[''+id]);
        });
      },

      handleSearch(e) {
        e.preventDefault();
        window.handleSearch(this.filter);
      },

      initialize() {
        document.title += ' - Search for "' + this.filter + '"';
        if (this.filter) this.fetchData();
      },
    }
  }

  result.transactions = function() {
    return {
      txns: undefined,

      initialize() {
        document.title += ' - Market Transactions'
        sdeTypes().then(types => {
          this.types = types;
          return sdeStations();
        })
        .then(stations => {
          this.stations = stations;
          return  retrieve('/api/v1/user/transactions', 'error fetching wallet transactions')
        })
        .then(txns => {
          this.txns = txns.map(t => setOrderFields(t, this.types, this.stations));
        });
      },
    }
  }

  result.typeDetails = function() {
    let typeRE = new RegExp("/types/(.*)");
    let match = window.location.pathname.match(typeRE);

    return {
      group: {},
      type: undefined,
      typeID: match[1],
      info: undefined,
      favorite: false,
      station: undefined,
      system: { name: "" },

      toggleFavorite() {
        setFavorite(this.typeID, !this.favorite)
        .then(obj => {
          this.favorite = obj.favorite;
        });
      },

      get parentGroups() {
        const arr = [];
        var g = this.group
        while (g.parentID) {
          g = this.marketGroups.groups[g.parentID];
          arr.unshift(g);
        }
        return arr
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

        currentUser.then(user => {
          this.user = user;
          return sdeStations();
        })
        .then(stations => {
          this.station = stations[''+this.user.station_id];
          return sdeSystems();
        })
        .then(systems => {
          this.system = systems[''+this.station.system_id]
        })
        .then(() => {
          const params = new URLSearchParams();
          params.set("location_id", this.station.id);
          params.set("region_id", this.system.region_id);
          const url = '/api/v1/types/' + this.typeID + '?' + params.toString();
          return retrieve(url);
        })
        .then(obj => {
          obj.history = obj.history.map(d => {
            return {
              date: Date.parse(d.date),
              average: +d.average,
            }
          });
          this.info = obj;
          this.favorite = obj.favorite;
        });

        sdeTypes().then(types => {
          this.type = types[''+this.typeID];
          document.title += ' - ' + this.type.name;
          return sdeMarketGroups()
        })
        .then(marketGroups => {
          this.marketGroups = marketGroups
          this.group = marketGroups.groups[''+this.type.market_group_id];
        });
      }
    }
  }

  // Helper Functions

  window.avatarURL = function(id) {
    return id && 'https://images.evetech.net/characters/' + id + '/portrait?size=128';
  }

  window.hrefGroup = function(id) {
    return '/groups/' + id;
  }

  window.hrefType = function(id) {
    return '/types/' + id;
  },

  window.imgURL = function(type) {
    if (!type) return undefined;
    var imgType = 'icon';
    if (type.name.match(/(Blueprint|Formula)$/))
      imgType = 'bp';
    return 'https://images.evetech.net/types/' + type.id + '/' + imgType + '?size=128';
  }

  window.formatISK = function(amt) {
    return amt.toLocaleString('en-US', { maximumFractionDigits: 2, minimumFractionDigits: 2 });
  }

  window.formatNumber = function(amt) {
    return amt.toLocaleString('en-US', { maximumFractionDigits: 2 });
  }

  window.handleSearch = function(q) {
    window.location = '/search?q=' + q;
  }

  window.makeActiveCharacter = function(cid) {
    retrieve('/api/v1/user/characters/' + cid + '/activate', 'error activating user', {
      raw: true,
      method: 'POST',
    })
    .then(() => {
      location.reload(true);
    });
  }

  window.openTypeInEVE = function(typeID) {
    retrieve('/api/v1/types/'+typeID+'/openInGame', 'error making openInGame API call', {
      raw: true,
      method: 'POST',
    });
  }

  window.removeCharacter = function(cid) {
    retrieve('/api/v1/user/characters/' + cid, 'error deleting character', {
      raw: true,
      method: 'DELETE',
    })
    .then(() => {
      location.reload(true);
    });
  }

  window.setFavorite = function(typeID, val) {
    return retrieve('/api/v1/types/'+typeID+'/favorite', 'error setting favorite', {
      method: 'PUT',
      body: JSON.stringify({favorite: val}),
    });
  }

  return result;
})(document, window);
