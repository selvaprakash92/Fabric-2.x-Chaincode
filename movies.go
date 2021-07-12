package main
 
import (
    "fmt"
    "encoding/json"
    "strconv"
    "strings"
    "bytes"
    "time"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)
 
// Define the Smart Contract structure
type ChainCode struct {
    contractapi.Contract
}
 
const NEXT_SHOW_ID   = "NEXT_SHOW_ID"
const NEXT_TICKET_ID = "NEXT_TICKET_ID"
const SHOW   = "SHOW"
const TICKET = "TICKET"
const WINDOW = "WINDOW"
const SODA   = "SODA"

 
type Theatre struct {
    TheatreNo      int    `json:"theatreNo"`
    TheatreName    string `json:"theatreName"`
    Windows        int    `json:"windows, omitempty"`
    TicketsPerShow int    `json:"ticketsPerShow, omitempty"`
    ShowsDaily     int    `json:"showsDaily, omitempty"`
    SodaStock      int    `json:"sodaStock, omitempty"`
    Halls          int    `json:"halls, omitempty"`
    DocType        string `json:"docType"`
}
 
type Window struct {
    WindowNo    int    `json:"windowNo"`
    TicketsSold int    `json:"ticketsSold"`
    DocType     string `json:"docType"`
}
 
type Ticket struct {
    TicketNo        int     `json:"ticketNo"`
    Show            Show    `json:"show"`
    Window          Window  `json:"window"`
    Quantity        int     `json:"quantity,number"`
    Amount          float64 `json:"amount,string"`
    CouponNumber    string  `json:"couponNumber"`
    CouponAvailed   bool    `json:"couponAvailed"`
    ExchangeAvailed bool    `json:"exchangeAvailed"`
    DocType         string  `json:"docType"`
}
 
type Show struct {
    ShowID    int    `json:"showID"`
    Movie     string `json:"movie"`
    ShowSlot  string `json:"showSlot"`
    Quantity  int    `json:"quantity,number"`
    HallNo    int    `json:"hallNo"`
    TheatreNo int    `json:"theatreNo"`
    DocType   string `json:"docType"`
}
 
type Soda struct {
    Stock        int    `json:"stock"`
    TicketNo     int    `json:"ticketNo"`
    CouponNumber string `json:"couponNumber"`
    DocType      string `json:"docType"`
}
 
type Property struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}
 
type CreateShows struct {
    TheatreNo int    `json:"theatreNo"`
    Shows     []Show `json:"shows"`
}
 
// =========================================================================================
// The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
// Best practice is to have any Ledger initialization in separate function -- see initLedger()
// =========================================================================================
func (c *ChainCode) Init(ctx contractapi.TransactionContextInterface) error {
 
    _, err := set(ctx.GetStub(), NEXT_SHOW_ID, "0")
    _, err = set(ctx.GetStub(), NEXT_TICKET_ID, "0")
 
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    return nil
}
 
func (c *ChainCode) registerTheatre(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
    if len(args) != 1 {
        return "nill",fmt.Errorf("Incorrect number of arguments. Expecting 1")
    }
    var theatre Theatre
    if err := json.Unmarshal([]byte(args[0]), &theatre); err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    // Create unique theatre number & save theatre
    txnId := ctx.GetStub().GetTxID()
    var number int
    for _, c := range txnId {
        number = number + int(c)
    }
    theatre.TheatreNo = number
    theatre.DocType = "THEATRE"
    theatreAsBytes, _ := json.Marshal(theatre)
    err := ctx.GetStub().PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    // create windows for the theatre
    for i := 1; i <= theatre.Windows; i++ {
        var window Window
        window.WindowNo = i
        window.TicketsSold = 0
        window.DocType = WINDOW
        windowAsBytes, _ := json.Marshal(window)
        err := ctx.GetStub().PutState("WINDOW"+strconv.Itoa(i), windowAsBytes)
        if err != nil {
            return "nil", fmt.Errorf(err.Error())
        }
    }
 
    return []byte("MovieTheatre Number:" + strconv.Itoa(theatre.TheatreNo)), nil
}
 
func (c *ChainCode) createShow(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
 
    if len(args) != 1 {
        return "nil", fmt.Errorf("Incorrect number of arguments. Expecting 1")
    }
 
    var createShows CreateShows
 
    if err := json.Unmarshal([]byte(args[0]), &createShows); err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    showSeq, err := get(ctx.GetStub(), NEXT_SHOW_ID)
 
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    shows := createShows.Shows
    var theatre Theatre
    theatreBytes, err := ctx.GetStub().GetState("THEATRE" + strconv.Itoa(createShows.TheatreNo))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    json.Unmarshal(theatreBytes, &theatre)
 
    if len(shows) > theatre.Halls {
        return "nil", fmt.Errorf("Number of Movies cannot exceed" + strconv.Itoa(theatre.Halls))
    }
    for _, show := range shows {
 
        for i := 1; i <= theatre.ShowsDaily; i++ {
            showSeq = showSeq + 1
            show.ShowID = +showSeq
            show.ShowSlot = strconv.Itoa(i)
            show.Quantity = theatre.TicketsPerShow
            show.TheatreNo = theatre.TheatreNo
            show.DocType = SHOW
            showAsBytes, _ := json.Marshal(show)
            err = ctx.GetStub().PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
            if err != nil {
                return "nil", fmt.Errorf(err.Error())
            }
        }
    }
    _, err = set(ctx.GetStub(), NEXT_SHOW_ID, strconv.Itoa(showSeq))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    return []byte(ctx.GetStub().GetTxID()), nil
}
 
func (c *ChainCode) purchaseTicket(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
    if len(args) != 1 {
        return "nil", fmt.Errorf("Incorrect number of arguments. Expecting 1")
    }
 
    var ticket Ticket
 
    if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    ticketSeq, err := get(ctx.GetStub(), NEXT_TICKET_ID)
 
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    showBytes, err := ctx.GetStub().GetState("SHOW" + strconv.Itoa(ticket.Show.ShowID))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    var show Show
    json.Unmarshal(showBytes, &show)
 
    windowBytes, err := ctx.GetStub().GetState("WINDOW" + strconv.Itoa(ticket.Window.WindowNo))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    var window Window
    json.Unmarshal(windowBytes, &window)
    // check the show for number of seats remaining
    if show.Quantity < 0 || show.Quantity-ticket.Quantity < 0 {
        return "nil", fmt.Errorf("Seats Full for the requested show or Not enough seats as requested. Available:" + strconv.Itoa(show.Quantity))
    }
 
    show.Quantity = show.Quantity - ticket.Quantity
    window.TicketsSold = window.TicketsSold + ticket.Quantity
    ticketSeq = ticketSeq + 1
    ticket.TicketNo = ticketSeq
    ticket.Show = show
    ticket.Window = window
    ticket.DocType = TICKET
 
    showAsBytes, _ := json.Marshal(show)
    err = ctx.GetStub().PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    windowAsBytes, _ := json.Marshal(window)
    err = ctx.GetStub().PutState("WINDOW"+strconv.Itoa(window.WindowNo), windowAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    _, err = set(ctx.GetStub(), NEXT_TICKET_ID, strconv.Itoa(ticketSeq))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    ticketAsBytes, _ := json.Marshal(ticket)
    err = ctx.GetStub().PutState("TICKET"+strconv.Itoa(ticketSeq), ticketAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    return []byte(ctx.GetStub().GetTxID()), nil
}
 
// Issue coupon for the waterbottle and popcorn also for the soda exchange
func (c *ChainCode) issueCoupon(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
 
    if len(args) != 1 {
        return "nil", fmt.Errorf("Incorrect number of arguments. Expecting 1")
    }
 
    var ticket Ticket
 
    if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    ticketBytes, err := ctx.GetStub().GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    json.Unmarshal(ticketBytes, &ticket)
 
    if ticket.CouponAvailed {
        return "nil", fmt.Errorf("Coupon Availed Already")
    }
 
    txnId := ctx.GetStub().GetTxID()
    var number int
    for _, c := range txnId {
        number = number + int(c)
    }
    ticket.CouponNumber = strconv.Itoa(number)
    ticket.CouponAvailed = true
    ticket.ExchangeAvailed = false
    ticketAsBytes, _ := json.Marshal(ticket)
    err = ctx.GetStub().PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    return []byte("Coupon Number:" + ticket.CouponNumber), nil
}
 
func (c *ChainCode) availExchange(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
    if len(args) != 1 {
        return "nil", fmt.Errorf("Incorrect number of arguments. Expecting 1")
    }
 
    var ticket Ticket
 
    if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    ticketBytes, err := ctx.GetStub().GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    json.Unmarshal(ticketBytes, &ticket)
 
    if ticket.ExchangeAvailed {
        return "nil", fmt.Errorf("Exchange Availed Already")
    }
    // check if even number for the eligible soda exchange
    couponNo, err := strconv.Atoi(ticket.CouponNumber)
    if err != nil {
        return "nil", fmt.Errorf("Ticket Not eligible for exchange")
    }
    if couponNo%2 != 0 {
        return "nil", fmt.Errorf("Ticket Not eligible for exchange")
    }
    ticket.ExchangeAvailed = true
    ticketAsBytes, _ := json.Marshal(ticket)
    err = ctx.GetStub().PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    var theatre Theatre
    theatreBytes, err := ctx.GetStub().GetState("THEATRE" + strconv.Itoa(ticket.Show.TheatreNo))
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
    json.Unmarshal(theatreBytes, &theatre)
 
    theatre.SodaStock = theatre.SodaStock - ticket.Quantity
 
    theatreAsBytes, _ := json.Marshal(theatre)
    err = ctx.GetStub().PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    return []byte(ctx.GetStub().GetTxID()), nil
}
 
func (c *ChainCode) queryByString(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {
 
    queryString := args[0]
    queryResults, err := getQueryResultForQueryString(ctx.GetStub(), queryString)
 
    if err != nil {
        return "nil", fmt.Errorf(err.Error())
    }
 
    return queryResults, nil
}
 
// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {
 
    fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
 
    resultsIterator, err := stub.GetQueryResult(queryString)
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()
 
    // buffer is a JSON array containing QueryRecords
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }
        // Add a comma before array members, suppress it for the first array member
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
        }
        buffer.WriteString(string(queryResponse.Value))
        bArrayMemberAlreadyWritten = true
    }
    buffer.WriteString("]")
 
    fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
 
    return buffer.Bytes(), nil
}
 
func get(ctx contractapi.TransactionContextInterface, key string) (int, error) {
    if key == "" {
        return 0, fmt.Errorf("Incorrect arguments. Expecting a key")
    }
    value, err := ctx.GetStub().GetState(key)
    if err != nil {
        return 0, fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
    }
    if value == nil {
        return 0, fmt.Errorf("Asset not found: %s", key)
    }
    var property Property
    json.Unmarshal(value, &property)
    i, err := strconv.Atoi(property.Value)
    if err != nil {
        return 0, fmt.Errorf("Failed to get next sequence number", err)
    }
    return i, nil
}
 
func set(ctx contractapi.TransactionContextInterface, key string, value string) (string, error) {
 
    var property Property
    property.Key = key
    property.Value = value
 
    propertyAsBytes, _ := json.Marshal(property)
    err := ctx.GetStub().PutState(key, propertyAsBytes)
    if err != nil {
        return "", fmt.Errorf(err.Error())
    }
    return value, nil
}
 
func main() {
    movieChaincode, err := contractapi.NewChaincode(&main.Chaincode{})
    if err != nil {
        log.Panicf("Error creating movieChaincode chaincode: %v", err)
    }
 
    if err := movieChaincode.Start(); err != nil {
        log.Panicf("Error starting movieChaincode chaincode: %v", err)
    }
}

