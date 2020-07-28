describe('evetools', function() {
  it('shows the branding', function() {
    expect(document.querySelectorAll('.branding').length).toEqual(1);
  });
});
