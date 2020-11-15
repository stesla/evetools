renderChart = function(history, height, width) {
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
