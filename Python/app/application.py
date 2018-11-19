from flask import Flask, render_template, request, jsonify
from datetime import datetime
import requests, json, time, os

app = Flask(__name__)

# use ENV VARIABLE / SECRET
# apikey = 'AIzaSyDRtANEJbPszXVTopPeo3s-FJkuHo0B_Qc'
apikey = os.environ['APIKEY']

if apikey == "":
  exit()

@app.route('/', methods=['GET', 'POST'])
def index():
    if request.method == "GET":
        return render_template("index.html")
    if request.method == "POST":
      city = request.form.get("city")

      results = gettime(city)
      return render_template("time.html", hour=results['hour'], date=results['date'], name=results['name'])

@app.route("/api/city=<string:city>")
def api(city):
  # /api/city= returns a JSON with local time, date and city's name 
  city = city.replace("+", " ")
  results = gettime(city)
  return jsonify(results)


def gettime(city):
  info = {}
  # query API for latitude and longitude 
  url = "https://maps.googleapis.com/maps/api/geocode/json?address=" + city + "&language=en&key=" + apikey
  r = requests.get(url)

  # unserialize json
  res = json.loads(r.content)
  # check if response is 200 ! Change, check for error (dict)
  if res['status'] == "OK":        
    # get latitude
    lat = res['results'][0]['geometry']['location']['lat']
    # get longitude
    lng = res['results'][0]['geometry']['location']['lng']
    # get name of city
    name = res['results'][0]['formatted_address']

    # convert everything to string
    unixutcnow = int(time.time())
    lat = str(lat)
    lng = str(lng)
    unixutcnow = str(unixutcnow)

    # query API for timezone 
    t = requests.get("https://maps.googleapis.com/maps/api/timezone/json?location=" + lat + "," + lng + "&timestamp=" + unixutcnow + "&key=" + apikey)
    # unserialize json
    tres = json.loads(t.content)
    if tres['status'] == "OK":
      # get offset (seconds between current timezone and UTC)
      dstoffset = tres['dstOffset']
      rawoffset =  tres['rawOffset']
      intunixutcnow = int(unixutcnow)

      # calculate local time (city)
      timeunix = intunixutcnow + ((rawoffset) + (dstoffset))

      # store information on a dict 
      info['hour'] = datetime.utcfromtimestamp(timeunix).strftime('%H:%M')
      info['date'] = datetime.utcfromtimestamp(timeunix).strftime('%d/%m/%Y')
      info['name'] = name

      # TO DO: create info[day] / [month] and [year]. create a dict with months, pass to html
      # november for eg. dont change [date] (API) 

      return info
      #return render_template("time.html", hour=hour, date=date, name=name)
    else:
      return "API error, unable to find timezone"

  else:
    return "API error, unable to find city"



if __name__ == '__main__':
    app.run(debug=True,host='0.0.0.0')
  