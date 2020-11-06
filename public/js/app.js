var evetools = {}

evetools.globalState = function() {
  return {
    avatarMenuOpen: false,
    loggedIn: false,
    user: null,

    get currentView() {
      if (!this.loggedIn) {
        return 'login';
      }

      let path = window.location.pathname;
      
      if (path.startsWith('/type/')) {
        return 'typeDetails';
      } else {
        return 'search';
      }
    },

    fetchCurrentUser() {
      fetch('/api/v1/currentUser').
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching current user');
          }
          return resp.json();
        }).
        then(user => {
          this.user = user
          this.user.avatarURL = 'https://images.evetech.net/characters/' + user.characterID + '/portrait?tenant=tranquility&size=128';
          this.loggedIn = true;
        }).catch(() => {});
    },

    handleEscape(e) {
      if (e.key === 'Esc' || e.key === 'Escape') {
        this.avatarMenuOpen = false;
      }
    }
  }
}

evetools.marketTypes = function() {
  const urlParams = new URLSearchParams(window.location.search);
  return {
    filter: urlParams.get('q'),
    marketTypes: [],

    fetchMarketTypes() {
      fetch('/api/v1/typeSearch/' + this.filter).
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching market types');
          }
          return resp.json();
        }).
        then(types => {
          this.marketTypes = types;
        });
    },

    href(id) {
      return '/type/' + id;
    },

    initMarketTypes() {
      if (this.filter)
        this.fetchMarketTypes();
    },

    handleSearch(e) {
      e.preventDefault();
      window.location = '?q=' + this.filter;
    },
  }
}

evetools.typeInfo = function() {
  let typeRE = new RegExp("/type/(.*)");
  let match = window.location.pathname.match(typeRE);
  let typeID = match[1];

  return {
    type: undefined,
    marketInfo: undefined,

    fetchTypeDetails() {
      fetch('/api/v1/types/' + typeID).
        then(resp => {
          if(!resp.ok) {
            throw new Error('error fetching type info');
          }
          return resp.json();
        }).
        then(type => {
          this.type = type;
        });

      fetch('/api/v1/types/' + typeID + '/marketInfo').
        then(resp => {
          if (!resp.ok) {
            throw new Error('error fetching market info');
          }
          return resp.json();
        }).
        then(obj => {
          obj.history = obj.history.map(d => {
            return {
              date: Date.parse(d.date),
              average: +d.average,
            }
          });
          this.marketInfo = obj;
        });

      const observer = new MutationObserver(() => {
        let div = document.getElementById("chart");
        if (div) {
          renderChart(this.marketInfo.history, div.clientWidth);
          observer.disconnect();
        }
      });
      observer.observe(document.querySelector('main'), { childList: true, subtree: true });
    },
  }
}

window.imgURL = function(type) {
  if (type) {
    return 'https://images.evetech.net/types/' + type.id + '/icon?tenant=tranquility&size=128';
  } else {
    return undefined;
  }
}

window.formatNumber = function(amt) {
  return amt.toLocaleString('en-US', { maximumFractionDigits: 2 });
}

window.renderChart = function(data, width) {
  var margin = {top: 20, right: 30, bottom: 0, left: 70};
  var height = 400;

  const bisect = function(mx) {
    const date = x.invert(mx);
    const index = d3.bisector(d => d.date).left(data, date, 1);
    const a = data[index - 1];
    const b = data[index];
    return b && (date - Date.parse(a.date) > Date.parse(b.date) - date) ? b : a;
  };

  const formatDate = function(date) {
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      timeZone: 'UTC'
    });
  }

  const formatValue = function(value) {
    return value.toLocaleString('en-US', {
      style: "currency",
      currency: "ISK"
    });
  }

  var line = d3.line()
    .curve(d3.curveStep)
    .defined(d => !isNaN(+d.average))
    .x(d => x(d.date))
    .y(d => y(d.average));

  var yAxis = g => g
    .attr('transform', `translate(${margin.left},0)`)
    .call(d3.axisLeft(y))
    .call(g => g.select('.domain').remove())
    .call(g => g.select('.tick:last-of-type text').clone()
        .attr('x', 3)
        .attr('text-anchor', 'start')
        .attr('font-weight', 'bold')
        .text(data.y));

  var xAxis = g => g
    .attr("transform", `translate(0,${height - margin.bottom})`)
    .call(d3.axisBottom(x).ticks(3).tickSizeOuter(0));

  var y = d3.scaleLinear()
          .domain([0, d3.max(data, d => d.average)]).nice()
          .range([height - margin.bottom, margin.top]);
  
  var x = d3.scaleUtc()
          .domain(d3.extent(data, d => d.date))
          .range([margin.left, width - margin.right]);

  var callout = function(g, value) {
    if (!value) return g.style('display', 'none');

    g.style('display', null)
     .style('pointer-events', 'none')
     .style('font', '10px sans-serif');

    const path = g.selectAll('path')
      .data([null])
      .join('path')
        .attr('fill', 'white')
        .attr('stroke', 'black');

    const text = g.selectAll('text')
      .data([null])
      .join('text')
      .call(text => text
        .selectAll('tspan')
        .data((value + '').split(/\n/))
        .join('tspan')
          .attr('x', 0)
          .attr('y', (d, i) => `${i * 1.1}em`)
          .style('font-weight', (_, i) => i ? null : 'bold')
          .text(d => d));

    const {x, y, width: w, height: h} = text.node().getBBox();

    text.attr('transform', `translate(${-w / 2},${15 - y})`);
    path.attr('d', `M${-w / 2 - 10},5H-5l5,-5l5,5H${w / 2 + 10}v${h + 20}h-${w + 20}z`);
  };

  const svg = d3.select("#chart").append("svg")
    .attr('width', width + margin.left + margin.right)
    .attr('height', height + margin.top + margin.bottom);

  svg.append('g').call(xAxis);

  svg.append('g').call(yAxis);

  svg.append('path')
    .datum(data)
    .attr('fill', 'none')
    .attr('stroke', 'steelblue')
    .attr('stroke-width', 1.5)
    .attr('stroke-linejoin', 'round')
    .attr('stroke-linecap', 'round')
    .attr('d', line);

  const tooltip = svg.append('g');

  svg.on('touchmove mousemove', function(event) {
    const day = bisect(d3.pointer(event, this)[0]);
    const date = new Date(day.date);
    const value = day.average;

    tooltip
      .attr("transform", `translate(${x(date)},${y(value)})`)
      .call(callout, `${formatValue(value)}
${formatDate(date)}`);
  });

  svg.on("touchend mouseleave", () => tooltip.call(callout, null));
}
