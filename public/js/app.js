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
  var sdeTypes = retrieve('/data/types.json', 'error fetching sde types');
  var sdeMarketGroups = retrieve('/data/marketGroups.json', 'error fetching sde market groups');
  var sdeStations = retrieve('/data/stations.json', 'error fetching sde stations');

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
        sdeMarketGroups.then(data => {
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
        if (!this.group)
          return [];

        if (this.group.groups) {
          return this.group.groups.map(id => {
            let g = this.marketGroups.groups[''+id];
            g.isGroup = true;
            return g
          }).sort(byName);
          return []
        } else if (this.group.types) {
          return this.group.types.map(id => {
            let t = this.types[''+id];
            t.isType = true;
            return t;
          }).sort(byName);
        }
      },

      initialize() {
        sdeMarketGroups.then(data => {
          this.marketGroups = data;
          this.group = data.groups[''+this.groupID];
          this.parent = data.groups[''+this.group.parent_id];
          document.title += " - " + this.group.name;
        });

        sdeTypes.then(types => {
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
          this.station = user.station;
          this.walletBalance = user.wallet_balance;
        })
        .then(() => {
          return retrieve('/api/v1/user/orders', 'error fetching market orders');
        })
        .then(orders => {
          this.buyTotal = orders.buy.reduce((a, o) => a + o.escrow, 0);
          this.sellTotal = orders.sell.reduce((a, x) => a + x.volume_remain * x.price, 0);
        });

        sdeTypes.then(types => {
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
        sdeStations.then(stations => {
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
        let station = this.stations[this.stationName];
        retrieve('/api/v1/user/station', 'error saving station', {
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
  }

  function setTypeFromID(o, types) {
    let type = types[''+o.type_id]
    o.type = type
    return o
  }

  function byTypeName(a, b) {
    return a.type.name < b.type.name ? -1 : 1;
  }

  result.history = function() {
    return {
      initialize() {
        document.title += ' - Market Order History'
        sdeTypes.then(types => {
          this.types = types;
          return retrieve('/api/v1/user/history?days=30', 'error fetching history');
        })
        .then(data => {
          this.orders = {
            buy: data.buy.map(o => setTypeFromID(o, this.types)),
            sell: data.sell.map(o => setTypeFromID(o, this.types)),
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
        sdeTypes.then(types => {
          this.types = types;
          return retrieve('/api/v1/user/orders', 'error fetching orders');
        })
        .then(data => {
          this.orders = {
            buy: data.buy.map(o => setTypeFromID(o, this.types)),
            sell: data.sell.map(o => setTypeFromID(o, this.types)),
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
        sdeTypes.then(types => {
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
        sdeTypes.then(types => {
          this.types = types;
          return  retrieve('/api/v1/user/transactions', 'error fetching wallet transactions')
        })
        .then(txns => {
          this.txns = txns.map(t => setTypeFromID(t, this.types));
        });
      },
    }
  }

  result.typeDetails = function() {
    let typeRE = new RegExp("/types/(.*)");
    let match = window.location.pathname.match(typeRE);

    return {
      group: undefined,
      type: undefined,
      typeID: match[1],
      info: undefined,
      favorite: false,
      system: { name: "" },

      toggleFavorite() {
        setFavorite(this.typeID, !this.favorite)
        .then(obj => {
          this.favorite = obj.favorite;
        });
      },

      get parentGroups() {
        if (!this.group) return [];
        const arr = [];
        var g = this.group
        while (g.parentID) {
          g = this.marketGroups.groups[g.parentID];
          arr.unshift(g);
        }
        return arr
      },

      initialize() {
        currentUser.then(user =>
          this.system = user.station.system
        );

        const observer = new MutationObserver(() => {
          let div = document.getElementById("chart");
          if (div) {
            observer.disconnect();
            renderChart(this.info.history, 400, div.clientWidth);
          }
        });
        observer.observe(document.querySelector('main'), { childList: true, subtree: true });

        retrieve('/api/v1/types/'+ this.typeID, 'error fetching type details')
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

        sdeTypes.then(types => {
          this.type = types[''+this.typeID];
          document.title += ' - ' + this.type.name;
          return sdeMarketGroups
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

  window.renderChart = function(history, height, width) {
    const bollinger = function(values, N, K) {
      let i = 0;
      let sum = 0;
      let sum2 = 0;
      const bands = K.map(() => new Float64Array(values.length).fill(NaN));
      for (let n = Math.min(N - 1, values.length); i < n; ++i) {
        const value = values[i];
        sum += value, sum2 += value ** 2;
      }
      for (let n = values.length, m = bands.length; i < n; ++i) {
        const value = values[i];
        sum += value, sum2 += value ** 2;
        const mean = sum / N;
        const deviation = Math.sqrt((sum2 - sum ** 2/ N) / (N - 1));
        for (let j = 0; j < K.length; ++j) {
          bands[j][i] = mean + deviation * K[j];
        }
        const value0 = values[i - N + 1];
        sum -= value0, sum2 -= value0 ** 2;
      }
      return bands;
    };

    const margin = {top: 20, right: 30, bottom: 20, left: 70};

    const svg = d3.select("#chart").append("svg")
      .attr('viewBox', `0 0 ${width} ${height}`);

    const values = Float64Array.from(history, d => d.average);

    const y = d3.scaleLinear()
            .domain(d3.extent(values)).nice()
            .range([height - margin.bottom, margin.top]);

    const yAxis = g => g
      .attr('transform', `translate(${margin.left},0)`)
      .call(d3.axisLeft(y).tickValues(d3.ticks(...y.domain(), 10)).tickFormat(d => d))
      .call(g => g.select('.domain').remove())
      .call(g => g.selectAll('.tick line').clone()
          .attr('x2', width - margin.left - margin.right)
          .attr('stroke-opacity', 0.1))
      .call(g => g.select('.tick:last-of-type text').clone()
          .attr('x', 3)
          .attr('text-anchor', 'start')
          .attr('font-weight', 'bold')
          .text(history.y));

    svg.append('g').call(yAxis);

    const x = d3.scaleUtc()
            .domain(d3.extent(history, d => d.date))
            .range([margin.left, width - margin.right]);

    const xAxis = g => g
      .attr("transform", `translate(0,${height - margin.bottom})`)
      .call(d3.axisBottom(x).ticks(3).tickSizeOuter(0));

    svg.append('g').call(xAxis);

    const line = d3.line()
      .defined(d => !isNaN(d))
      .x((d, i) => x(history[i].date))
      .y(y);

    const N = 7 // days
    const K = 2  // standard deviations

    const data = [
      values,
      ...bollinger(values, 7, [0]),
      ...bollinger(values, 30, [0]),
      ...bollinger(values, 60, [0]),
    ]

    const categories = ['raw', '7-day', '30-day', '60-day'];
    const colors = d3.scaleOrdinal(categories, ['#ddd', 'green', 'red', 'blue']);

    // make the 60-day line a little thicker
    const widths = (i) => [1, 1, 1, 2][i];

    svg.append('g')
        .attr('fill', 'none')
        .attr('stroke-linejoin', 'round')
        .attr('stroke-linecap', 'round')
      .selectAll('path')
      .data(data)
      .join("path")
        .attr("stroke", (d, i) => colors(i))
        .attr('stroke-width', (d, i) => widths(i))
        .attr('d', line);

    const legend = svg => {
      const g = svg
          .attr("transform", `translate(${width},10)`)
          .attr("text-anchor", "end")
          .attr("font-family", "sans-serif")
          .attr("font-size", 10)
        .selectAll("g")
        .data(categories)
        .join("g")
          .attr("transform", (d, i) => `translate(0,${i * 20})`);

      g.append("rect")
          .attr("x", -48)
          .attr("width", 18)
          .attr("height", 18)
          .attr("fill", colors);

      g.append("text")
          .attr("x", -55)
          .attr("y", 9.5)
          .attr("dy", "0.35em")
          .text(d => d);
    }

    svg.append("g").call(legend);
  }

  return result;
})(document, window);
