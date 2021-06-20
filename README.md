# rn2483

Go library for interfacing with [RN2483][RN2483 product page] and [RN2903][RN2903 product page] LoRa modules via their
serial interface, particularly [LoStik][LoStik product page] USB devices.

## References

* [LoStik devices][LoStik product page] used to test this library
    * [LoStik GitHub repo][LoStik github repo] - provides a nice introduction to interacting with RN2483 and RN2903
      modules
* [RN2483 Command Reference][RN2483 command reference]
* [RN2903 Command Reference][RN2903 command reference]

## Alternatives

* https://github.com/sagneessens/RN2483 - a more fully featured library, but is implemented using a global singleton
  restricting you to one device per application and has limited error handling/reporting.

[LoStik product page]: https://ronoth.com/products/lostik

[LoStik github repo]: https://github.com/ronoth/LoStik

[RN2483 product page]: https://www.microchip.com/wwwproducts/en/RN2483

[RN2903 product page]: https://www.microchip.com/wwwproducts/en/RN2903

[RN2483 command reference]: http://ww1.microchip.com/downloads/en/DeviceDoc/40001784B.pdf

[RN2903 command reference]: http://ww1.microchip.com/downloads/en/DeviceDoc/40001811A.pdf
