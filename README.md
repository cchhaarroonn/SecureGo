# SecureGo

SecureGo is a simple RESTful API that allows users to create accounts, remove accounts, and create license keys.

## Requirements

- Go 1.16 or higher
- MongoDB 4.4 or higher

## Installation

1. Clone the repository
2. Run go build to build the binary
3. Run the binary with the following command: ./securego

## Configuration

SecureGo requires the MongoDB URI to be set. This can be done by changing "YOUR DATABASE CONNECT URI" on line 26 to your URI for connecting to database

## Usage

|                     **Route**                        |                                  **Description**                              | **Methods** |
|------------------------------------------------------|-------------------------------------------------------------------------------|-------------|
| /securego/createUser/:username/:password/:license    | Create user account by username, password, license key                        | POST        |
| /securego/removeUser/:username/:license              | Remove user account by username and license                                   | POST        |
| /securego/removeUser/:username                       | Remove user account by username                                               | POST        |
| /securego/createLicense                              | Create license key autotmatically with random letters                         | POST        |
| /securego/createLicense/:name                        | Create specific licence key                                                   | POST        |
| /securego/removeLicense/:name                        | Remove license from the database, also remove all users who had that license  | POST        |
| /securego/checkLicense/:name                         | Check license key by the name that is specified                               | GET         |
| /securego/getLicenses/                               | Get all licenses in database                                                  | GET         |

## Contribution

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.
