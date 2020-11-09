evetools = (function(document, window, undefined) {
  var result = {};
  var promisedData;

  result.globalState = function() {
    return {
      avatarMenuOpen: false,
      loggedIn: false,
      navOpen: false,
      user: { characterName: "", characterID: 0},

      get currentView() {
        if (!this.loggedIn) {
          return 'login';
        }

        let path = window.location.pathname;

        if (path.startsWith('/groups/'))
          return 'groupDetails';

        if (path.startsWith('/browse'))
          return 'browse';

        if (path.startsWith('/search'))
          return 'search';

        if (path.startsWith('/types/'))
          return 'typeDetails';

        return 'index';
      },

      initialize() {
        fetch('/api/v1/currentUser')
        .then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching current user');
          }
          return resp.json();
        })
        .then(user => {
          this.user = user
          this.user.avatarURL = 'https://images.evetech.net/characters/' + user.characterID + '/portrait?size=128';
          this.loggedIn = true;
        })
        .catch(() => {})
        .then(() => {
          const url = '/views/'+this.currentView+'.html';
          return fetch(url);
        })
        .then(resp => {
          if(!resp.ok) {
            throw new Error('error fetching view');
          }
          return resp.text();
        })
        .then(html => {
          const parser = new DOMParser();
          const elt = parser.parseFromString(html, 'text/html').querySelector('main');
          const slot = document.querySelector('main');
          slot.parentNode.replaceChild(elt, slot);
        });


        promisedData = fetch('/data/static.json')
        .then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching static data');
          }
          return resp.json();
        });
      },

      handleEscape(e) {
        if (e.key === 'Esc' || e.key === 'Escape') {
          this.avatarMenuOpen = false;
        }
      }
    }
  }

  // Views

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
        );
      },

      initialize() {
        promisedData.then(data => {
          this.data = data
        });
      }
    }
  }

  result.groupDetails = function() {
    let typeRE = new RegExp("/groups/(.*)");
    let match = window.location.pathname.match(typeRE);

    return {
      group: { name: "", groups: [] },
      groupID: match[1],
      data: { root: [] },
      filter: "",
      parent: { name: "" },

      get children() {
        if (!this.group)
          return [];

        if (this.group.groups) {
          return this.group.groups.map(id => {
            let g = this.data.groups[''+id];
            g.isGroup = true;
            return g
          });
          return []
        } else if (this.group.types) {
          return this.group.types.map(id => {
            let t = this.data.types[''+id];
            t.isType = true;
            return t;
          });
        }
      },

      initialize() {
        promisedData.then(data => {
          this.data = data;
          this.group = data.groups[''+this.groupID];
          this.parent = data.groups[''+this.group.parentID];
        });
      },
    }
  }

  result.index = function() {
    return {
      data: undefined,
      favorites: [],

      initialize() {
        promisedData
        .then(data => {
          this.data = data
          return fetch('/api/v1/types/favorites')
        })
        .then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching favorites');
          }
          return resp.json();
        })
        .then(types => {
          types.forEach(t => {
            this.favorites.push(this.data.types[''+t.id])
          })
        })
      },

      openTypeInEVE(typeID) {
        fetch('/api/v1/types/'+typeID+'/openInGame', {
          method: 'POST',
        })
        .then(resp => {
          // It will return 204 No Content, so this is all we need.
          if (!resp.ok) {
            throw new Error("error making openInGame API call");
          }
        });
      },
    }
  }

  result.search = function() {
    const urlParams = new URLSearchParams(window.location.search);
    return {
      data: undefined,
      filter: urlParams.get('q'),
      marketTypes: [],

      fetchData() {
        fetch('/api/v1/types/search/' + this.filter)
        .then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching search results');
          }
          return resp.json();
        })
        .then(ids => {
          this.marketTypes = ids.map(id => this.data.types[''+id])
        });
      },

      handleSearch(e) {
        e.preventDefault();
        window.handleSearch(this.filter);
      },

      initialize() {
        promisedData.then(data => {
          this.data = data;
        })
        .then(() => {
          if (this.filter) this.fetchData();
        });
      },
    }
  }

  result.typeDetails = function() {
    let typeRE = new RegExp("/types/(.*)");
    let match = window.location.pathname.match(typeRE);

    return {
      data: undefined,
      group: undefined,
      type: undefined,
      typeID: match[1],
      info: undefined,
      favorite: false,

      toggleFavorite() {
        fetch('/api/v1/types/details/'+this.typeID+'/favorite', {
          method: 'PUT',
          body: JSON.stringify({favorite: !this.favorite}),
        })
        .then(resp => {
          if (!resp.ok) {
            throw new Error("error setting favorite");
          }
          return resp.json();
        })
        .then(obj => {
          this.favorite = obj.favorite;
        });
      },

      get parentGroups() {
        if (!this.group) return [];
        const arr = [];
        var g = this.group
        while (g.parentID) {
          g = this.data.groups[g.parentID];
          arr.unshift(g);
        }
        return arr
      },

      fetchData() {
        fetch('/api/v1/types/details/'+ this.typeID)
        .then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching type details');
          }
          return resp.json();
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

        const observer = new MutationObserver(() => {
          let div = document.getElementById("chart");
          if (div) {
            observer.disconnect();
            renderChart(this.info.history, 400, div.clientWidth);
          }
        });
        observer.observe(document.querySelector('main'), { childList: true, subtree: true });
      },

      initialize() {
        promisedData.then(data => {
          this.data = data;
          this.type = data.types[''+this.typeID];
          this.group = data.groups[''+this.type.groupID];
        })
        .then(() => { this.fetchData() });
      }
    }
  }

  // Helper Functions

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

  window.formatNumber = function(amt) {
    return amt.toLocaleString('en-US', { maximumFractionDigits: 2 });
  }

  window.handleSearch = function(q) {
    window.location = '/search?q=' + q;
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
