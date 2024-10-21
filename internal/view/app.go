package view

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/common-nighthawk/go-figure"
	"github.com/gdamore/tcell/v2"
	"github.com/isan-rivkin/surf/internal/ui"
	"github.com/rivo/tview"
)

/*
------------------------
| context |		| logo |
------------------------
| 		search bar 	   |
------------------------
|       table          |
|                      |
------------------------
| bread crumbs |       |
------------------------
*/

type App struct {
	app               *tview.Application
	name              string
	version           string
	autocompleteTypes []string
}

func NewApp(name, version string, autocompleteTypes []string) *App {
	return &App{
		app:               tview.NewApplication(),
		name:              name,
		version:           version,
		autocompleteTypes: autocompleteTypes,
	}
}
func generateRandomDigits(n int) string {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	digits := "0123456789"
	result := make([]byte, n)

	for i := range result {
		result[i] = digits[rand.Intn(len(digits))] // Random digit
	}

	return string(result)
}

// Generate a random word with a random length
func generateRandomWord(minLen, maxLen int) string {
	letters := "abcdefghijklmnopqrstuvwxyz"
	wordLength := rand.Intn(maxLen-minLen+1) + minLen
	word := make([]byte, wordLength)

	for i := range word {
		word[i] = letters[rand.Intn(len(letters))]
	}

	return string(word)
}

// Generate a random sentence of length between minLen and maxLen characters
func generateRandomSentence(minLen, maxLen int) string {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	var sentence strings.Builder

	// Choose a random sentence length between minLen and maxLen
	sentenceLength := rand.Intn(maxLen-minLen+1) + minLen

	// Keep generating words until we reach the desired length
	for sentence.Len() < sentenceLength {
		word := generateRandomWord(3, 8) // Random word between 3 and 8 characters

		// Ensure we don't exceed the maximum length, accounting for spaces
		if sentence.Len()+len(word)+1 > sentenceLength {
			break
		}

		if sentence.Len() > 0 {
			sentence.WriteString(" ") // Add a space between words
		}
		sentence.WriteString(word)
	}

	return sentence.String()
}
func mockSecurityGroupResourcesTable() *tview.Table {
	t := tview.NewTable()
	t.SetBorders(false)
	rowsSize := 200
	cols := []string{"#", "ID", "Name", "Description", "Ingress Rules", "Egress Rules", "VpcId"}
	rows := [][]string{
		{"1", "sg-12345678987654322", "my-security-group", "Clusters Test Security group created in Python", "4", "1", "vpc-12345678987654321"},
	}
	for i := 0; i < rowsSize; i++ {
		idx := fmt.Sprintf("%d", i+2)
		id := fmt.Sprintf("sg-%s", generateRandomDigits(17))
		vpc := fmt.Sprintf("vpc-%s", generateRandomDigits(17))
		name := generateRandomWord(5, 20)
		description := generateRandomSentence(0, 100)
		egressRules := fmt.Sprintf("%d", rand.Intn(10))
		ingressRules := fmt.Sprintf("%d", rand.Intn(10))
		rows = append(rows, []string{idx, id, name, description, ingressRules, egressRules, vpc})
	}
	// update columns
	for i := 0; i < len(cols); i++ {
		t.SetCell(0, i, tview.NewTableCell(cols[i]).SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	}
	// update rows
	for i := 0; i < len(rows); i++ {
		for j := 0; j < len(rows[i]); j++ {
			t.SetCell(i+1, j, tview.NewTableCell(rows[i][j]).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
		}
	}
	t.SetFixed(1, 1)
	t.SetSelectable(true, false)
	t.Select(1, 0)
	t.SetSelectedFunc(func(row int, column int) {
		t.GetCell(row, column).SetTextColor(tcell.ColorRed)
	})
	return t
}

func (a *App) getPromptResourceTypes() []string {
	// const words = "ability,able,about,above,accept,according,account,across,act,action,activity,actually,add,address,administration,admit,adult,affect,after,again,against,age,agency,agent,ago,agree,agreement,ahead,air,all,allow,almost,alone,along,already,also,although,always,American,among,amount,analysis,and,animal,another,answer,any,anyone,anything,appear,apply,approach,area,argue,arm,around,arrive,art,article,artist,as,ask,assume,at,attack,attention,attorney,audience,author,authority,available,avoid,away,baby,back,bad,bag,ball,bank,bar,base,be,beat,beautiful,because,become,bed,before,begin,behavior,behind,believe,benefit,best,better,between,beyond,big,bill,billion,bit,black,blood,blue,board,body,book,born,both,box,boy,break,bring,brother,budget,build,building,business,but,buy,by,call,camera,campaign,can,cancer,candidate,capital,car,card,care,career,carry,case,catch,cause,cell,center,central,century,certain,certainly,chair,challenge,chance,change,character,charge,check,child,choice,choose,church,citizen,city,civil,claim,class,clear,clearly,close,coach,cold,collection,college,color,come,commercial,common,community,company,compare,computer,concern,condition,conference,Congress,consider,consumer,contain,continue,control,cost,could,country,couple,course,court,cover,create,crime,cultural,culture,cup,current,customer,cut,dark,data,daughter,day,dead,deal,death,debate,decade,decide,decision,deep,defense,degree,Democrat,democratic,describe,design,despite,detail,determine,develop,development,die,difference,different,difficult,dinner,direction,director,discover,discuss,discussion,disease,do,doctor,dog,door,down,draw,dream,drive,drop,drug,during,each,early,east,easy,eat,economic,economy,edge,education,effect,effort,eight,either,election,else,employee,end,energy,enjoy,enough,enter,entire,environment,environmental,especially,establish,even,evening,event,ever,every,everybody,everyone,everything,evidence,exactly,example,executive,exist,expect,experience,expert,explain,eye,face,fact,factor,fail,fall,family,far,fast,father,fear,federal,feel,feeling,few,field,fight,figure,fill,film,final,finally,financial,find,fine,finger,finish,fire,firm,first,fish,five,floor,fly,focus,follow,food,foot,for,force,foreign,forget,form,former,forward,four,free,friend,from,front,full,fund,future,game,garden,gas,general,generation,get,girl,give,glass,go,goal,good,government,great,green,ground,group,grow,growth,guess,gun,guy,hair,half,hand,hang,happen,happy,hard,have,he,head,health,hear,heart,heat,heavy,help,her,here,herself,high,him,himself,his,history,hit,hold,home,hope,hospital,hot,hotel,hour,house,how,however,huge,human,hundred,husband,idea,identify,if,image,imagine,impact,important,improve,in,include,including,increase,indeed,indicate,individual,industry,information,inside,instead,institution,interest,interesting,international,interview,into,investment,involve,issue,it,item,its,itself,job,join,just,keep,key,kid,kill,kind,kitchen,know,knowledge,land,language,large,last,late,later,laugh,law,lawyer,lay,lead,leader,learn,least,leave,left,leg,legal,less,let,letter,level,lie,life,light,like,likely,line,list,listen,little,live,local,long,look,lose,loss,lot,love,low,machine,magazine,main,maintain,major,majority,make,man,manage,management,manager,many,market,marriage,material,matter,may,maybe,me,mean,measure,media,medical,meet,meeting,member,memory,mention,message,method,middle,might,military,million,mind,minute,miss,mission,model,modern,moment,money,month,more,morning,most,mother,mouth,move,movement,movie,Mr,Mrs,much,music,must,my,myself,n't,name,nation,national,natural,nature,near,nearly,necessary,need,network,never,new,news,newspaper,next,nice,night,no,none,nor,north,not,note,nothing,notice,now,number,occur,of,off,offer,office,officer,official,often,oh,oil,ok,old,on,once,one,only,onto,open,operation,opportunity,option,or,order,organization,other,others,our,out,outside,over,own,owner,page,pain,painting,paper,parent,part,participant,particular,particularly,partner,party,pass,past,patient,pattern,pay,peace,people,per,perform,performance,perhaps,period,person,personal,phone,physical,pick,picture,piece,place,plan,plant,play,player,PM,point,police,policy,political,politics,poor,popular,population,position,positive,possible,power,practice,prepare,present,president,pressure,pretty,prevent,price,private,probably,problem,process,produce,product,production,professional,professor,program,project,property,protect,prove,provide,public,pull,purpose,push,put,quality,question,quickly,quite,race,radio,raise,range,rate,rather,reach,read,ready,real,reality,realize,really,reason,receive,recent,recently,recognize,record,red,reduce,reflect,region,relate,relationship,religious,remain,remember,remove,report,represent,Republican,require,research,resource,respond,response,responsibility,rest,result,return,reveal,rich,right,rise,risk,road,rock,role,room,rule,run,safe,same,save,say,scene,school,science,scientist,score,sea,season,seat,second,section,security,see,seek,seem,sell,send,senior,sense,series,serious,serve,service,set,seven,several,sex,sexual,shake,share,she,shoot,short,shot,should,shoulder,show,side,sign,significant,similar,simple,simply,since,sing,single,sister,sit,site,situation,six,size,skill,skin,small,smile,so,social,society,soldier,some,somebody,someone,something,sometimes,son,song,soon,sort,sound,source,south,southern,space,speak,special,specific,speech,spend,sport,spring,staff,stage,stand,standard,star,start,state,statement,station,stay,step,still,stock,stop,store,story,strategy,street,strong,structure,student,study,stuff,style,subject,success,successful,such,suddenly,suffer,suggest,summer,support,sure,surface,system,table,take,talk,task,tax,teach,teacher,team,technology,television,tell,ten,tend,term,test,than,thank,that,the,their,them,themselves,then,theory,there,these,they,thing,think,third,this,those,though,thought,thousand,threat,three,through,throughout,throw,thus,time,to,today,together,tonight,too,top,total,tough,toward,town,trade,traditional,training,travel,treat,treatment,tree,trial,trip,trouble,true,truth,try,turn,TV,two,type,under,understand,unit,until,up,upon,us,use,usually,value,various,very,victim,view,violence,visit,voice,vote,wait,walk,wall,want,war,watch,water,way,we,weapon,wear,week,weight,well,west,western,what,whatever,when,where,whether,which,while,white,who,whole,whom,whose,why,wide,wife,will,win,wind,window,wish,with,within,without,woman,wonder,word,work,worker,world,worry,would,write,writer,wrong,yard,yeah,year,yes,yet,you,young,your,yourself"
	// return strings.Split(words, ",")
	return a.autocompleteTypes
}

func (a *App) Init() error {
	// taking inspiration from https://github.com/rivo/tview/wiki/Postgres
	// Flexbox for layout and pages with stack

	// create context box
	contextTable := tview.NewTable().SetBorders(false)
	// mock default profile
	contextTable.SetCell(0, 0, tview.NewTableCell("Profile:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(0, 1, tview.NewTableCell("default").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	// mock sts get-caller-identity
	contextTable.SetCell(1, 0, tview.NewTableCell("Principal:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(1, 1, tview.NewTableCell("arn:aws:iam::123:role/Dev/MyRole").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	// mock region us-west-2
	contextTable.SetCell(2, 0, tview.NewTableCell("Region:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(2, 1, tview.NewTableCell("us-west-2").SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	// mock version
	contextTable.SetCell(2, 0, tview.NewTableCell("Surf Rev:").SetTextColor(tcell.ColorOrange).SetAlign(tview.AlignLeft))
	contextTable.SetCell(2, 1, tview.NewTableCell(a.version).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

	// add logo
	fig := figure.NewFigure(a.name, "starwars", true)
	title := fig.String()
	title = tview.TranslateANSI(title)

	logo := tview.NewTextView().SetText(title).SetTextColor(tcell.ColorOrange)
	flexHeader := tview.NewFlex().
		SetDirection(tview.FlexRowCSS).
		AddItem(contextTable, 0, 1, false).
		AddItem(logo, 50, 1, false)

	// add prompt
	prompt := tview.NewTextView()
	prompt.SetWordWrap(true)
	prompt.SetWrap(true)
	prompt.SetDynamicColors(true)
	prompt.SetBorder(true)
	prompt.SetBorderPadding(0, 0, 1, 1)
	prompt.SetTextColor(tcell.ColorPaleTurquoise)
	prompt.SetText("💡> AWS::EC2::SecurityGroup ")

	// input field can be implement sync and async https://github.com/rivo/tview/wiki/resourcesPrompt
	resourcesPrompt := tview.NewInputField().
		SetLabel("💡> ").
		// SetFieldWidth(30).
		SetDoneFunc(func(key tcell.Key) {
			fmt.Println("autocomplete done")
		})
	resourcesPrompt.SetBorder(true)
	resourcesPrompt.SetBorderPadding(0, 0, 1, 1)
	resourcesPrompt.SetFieldBackgroundColor(tcell.ColorBlack)
	// the autocomplete box popup of options that opens up
	resourcesPrompt.SetAutocompleteStyles(tcell.ColorBlack,
		tcell.StyleDefault.Foreground(tcell.ColorPaleTurquoise),
		tcell.StyleDefault.Foreground(tcell.ColorOrange),
	)
	resourcesPrompt.SetFieldTextColor(tcell.ColorPaleTurquoise)
	resourcesPrompt.SetBackgroundColor(tcell.ColorBlack)
	// resourcesPrompt.SetLabelColor(tcell.ColorGreen)

	resourcesPrompt.SetAutocompleteFunc(func(currentText string) (entries []string) {
		if len(currentText) == 0 {
			return
		}
		for _, resType := range a.getPromptResourceTypes() {
			if strings.HasPrefix(strings.ToLower(resType), strings.ToLower(currentText)) {
				entries = append(entries, resType)
			}
		}
		// in example its <= 1 but then it doesnt show partial with 1 match
		if len(entries) < 1 {
			entries = nil
		}
		return
	})

	resourcesPrompt.SetAutocompletedFunc(func(text string, index, source int) bool {
		if source != tview.AutocompletedNavigate {
			// once user hits "enter" selects here we can do stuff with selection like search it
			resourcesPrompt.SetText(text)
		}
		closeAutocompleteDropDown := source == tview.AutocompletedEnter || source == tview.AutocompletedClick
		return closeAutocompleteDropDown
	})

	// add resource description (box placeholder)
	sgTable := mockSecurityGroupResourcesTable()
	resources := sgTable
	// resources pane should be in pages and stack
	resourcesPages := tview.NewPages()
	resourcesPages.AddPage("resources", resources, true, true)

	// set layout using flexbox
	mainPage := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(flexHeader, 0, 1, false).
		AddItem(prompt, 3, 1, false).
		AddItem(resourcesPrompt, 3, 1, false).
		AddItem(resourcesPages, 0, 5, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("BreadCrumbs (3 rows)"), 3, 1, false)

	// create page
	pages := tview.NewPages()
	pages.AddPage("main", mainPage, true, true)
	a.app.SetRoot(pages, true)
	a.app.SetFocus(resources)

	// add keybindings
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// tcell.KeyCtrlBackslash
		switch event.Rune() {
		case ui.KeyColon:
			// resourcesPrompt.SetText(fmt.Sprintf("COLON!: name='%s' rune='%v' key='%d'", event.Name(), event.Rune(), event.Key()))
			a.app.SetFocus(resourcesPrompt)
			return event
		default:
			// resourcesPrompt.SetText(fmt.Sprintf("KEY: name='%s' rune='%v' key='%d'", event.Name(), event.Rune(), event.Key()))
		}
		switch event.Key() {
		case tcell.KeyEscape:
			a.app.SetFocus(resources)
			// case tcell.KeyCtrlC:
			// 	return nil
		}
		return event
	})
	return nil
}
func (a *App) Run() error {
	return a.app.Run()
}
