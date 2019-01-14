Feature: Hello World
  The user should see hello world

@HelloWorld001
Scenario Outline: Verify the user sees "<Message>"
  When this test case is executed, the user should see "<Message>"
  Examples:
    | Message |
    | Hello World |
    | Hallo Welt |
    | Bonjour le monde |
    | こんにちは世界|
