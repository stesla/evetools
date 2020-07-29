describe('evetools', function() {
  var main;
  beforeEach(function() {
    main = document.querySelector('section.main');
  });

  it('shows the branding', function() {
    expect(main.querySelectorAll('.branding').length).toEqual(1);
  });

  it('shows the landing view', function() {
    expect(main.querySelectorAll('.landing-view').length).toEqual(1);
  });
});
