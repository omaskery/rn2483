# rn2483

Go library for interfacing with [RN2483][RN2483 product page] and [RN2903][RN2903 product page] LoRa modules via their
serial interface, particularly [LoStik][LoStik product page] USB devices.

## Feature Completeness

- [x] Building blocks for implementing commands:
    - [x] `rn2483.(Device).Sendf` for sending commands
    - [x] `rn2483.(Device).ReadResponse` for reading responses
    - [x] `rn2483.CheckCommandResponse` for validating common command responses
    - [x] `rn2483.(Device).ExecuteCommand` as a common building block for commands
    - [x] `rn2483.(Device).ExecuteCommandChecked` and `rn2483.(Device).ExecuteCommandCheckedStrict` as a common building
      block for simple commands with easily validated responses
- [x] All `sys` commands
    - [ ] purposely excludes `sys eraseFW` as it seemed too dangerous to make convenient, easy to implement manually
      using the building blocks provided above
- [ ] `mac` commands have only been implemented where they facilitate accessing the `radio` commands:
    - [x] `mac pause`
- [ ] Basic `radio` commands have been implemented
    - [x] `radio tx` and `radio rx`
    - [x] generic `radio set <x> <y>` and `radio get <x>` commands
    - [x] `radio set pwr`
- [x] Simple fake implementation for local development and automated testing

## Todo

- Add documentation comments
- Add implementations for more commands

## References

* [LoStik devices][LoStik product page] used to test this library
    * [LoStik GitHub repo][LoStik github repo] - provides a nice introduction to interacting with RN2483 and RN2903
      modules
* [RN2483 Command Reference][RN2483 command reference]
* [RN2903 Command Reference][RN2903 command reference]

## Alternatives

* https://github.com/sagneessens/RN2483 - a more fully featured library, but is implemented using a global singleton
  restricting you to one device per application, depends on a specific serial port implementation, and has limited error
  handling/reporting.

[LoStik product page]: https://ronoth.com/products/lostik

[LoStik github repo]: https://github.com/ronoth/LoStik

[RN2483 product page]: https://www.microchip.com/wwwproducts/en/RN2483

[RN2903 product page]: https://www.microchip.com/wwwproducts/en/RN2903

[RN2483 command reference]: http://ww1.microchip.com/downloads/en/DeviceDoc/40001784B.pdf

[RN2903 command reference]: http://ww1.microchip.com/downloads/en/DeviceDoc/40001811A.pdf
