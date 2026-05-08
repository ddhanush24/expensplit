# Expensplit
A backend for expense-sharing application built usingGo and PostGreSQL

## Features
- User authentication and Profile
- Friend System
- Group System 
- Expense creation with custom naming and descriptions  
- Multiple split methods:  
-> Equal split  
-> Exact amount split  
-> Ratio/percentage split  
- Group and individual balance tracking  
- Pending/completed payment tracking  
- Expense editing 
- -Register Payment

## Setup Instructions
### Prerequisites
- Go 1.26+
- Postgre SQL

### Instructions
1. Clone the repo
run ` git clone github.com/ddhanush24/expensplit.git`
run`cd expensplit`

2. Install the dependencies 
run `go mod tidy`

3. Create a .env file of the folloiwing format
```
DB_URL=localhost
DB_NAME=yourdatabase_name
DB_PORT=5432
DB_USER=database_user_username
DB_PASS=database_user_password
SECRET_KEY=jwt_secret_key
```
4. Create a PostGre Database
```
CREATE DATABASE yourdatabase_name
```
(Replace yourdatabase_name with name of your choice both in the command and env file)

Dont have to create tables in the database. Running the server will automatically Create the required tables through `migrate.db`
5. Run Server
run `go run .`

Server will run at `http://localhost:8080`

## API Endpoints
### Authentication
|  Method|Endpoint|Description
|--|--|--|
| POST |`/Signup`  |Register New User |
| POST |`/Signin`  |Sign in User |

### Users
|  Method|Endpoint|Description
|--|--|--|
| GET|`/user`  |Returns User Data/Profile |
|DELETE |`/user`  |Deletes user if all Payments are settled |
| GET|`/dashbd`  |User Dashboard|




### Friends
|  Method|Endpoint|Description
|--|--|--|
| GET|`/friendlist`  |List of Friends |
| POST |`/friendp`  |Add Friend / Friend Request |
| GET|`/friendreqs`  | Get Friend requests|
| DELETE|`/friend`  |Remove Friend / Reject Friend Request |






### Expenses
|  Method|Endpoint|Description
|--|--|--|
| GET |`/xpense`  | Find all Expenses |
| POST |`/xpense`  |Create Expense |
| PATCH|`/xpense`  |Edit Created expense |
| DELETE|`/xpense`  |Remove or Delete Expense* |
|POST|`/payments`  |Register aPayment|
<sub>*Deletes expense only if has not been paid even partially, to maintain payment integrity</sub>

### Groups
|  Method|Endpoint|Description
|--|--|--|
| CREATE|`/Creategroup`  |Create a New Group |


## Expense split Logic
#### Amount Based SPlit
- Lets user specify amount share for each person in the group
- Remaining amount (which is not specified) is split among the remaining members equally.

```
For Example,
Total Amount = 120 and4 members
Rahuls Share = 40
Rohits Share = 20
Remaining 2 people pay 30 and 30 each
```

#### Equal Split
- Total Amount is split among the members equally.

```
For Example,
Total Amount = 120 and 4 members
Everyone pays 30 each
```
- Equal Split shares the same logic as amount basedaplit internally
#### Ratio & percentage Spilt
- Lets user specify amount share in ratio or percentage for each person in the group
- Requires share for every member to be mentioned

```
For Example,
Ratio 
Total Amount = 120 and 2 members
Rahul:Rohit = 2 Shares: 1 share
=> Rahul pays 2/3rd share or 80 and Rohit pays 40

Percentage
Total Amount = 120 and 2 members
Rahul = 35%
Rohit = 65%
=> Rahul pays 42 and Rohit pays 78

```

