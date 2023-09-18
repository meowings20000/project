from functools import partial
import json
import math
from PyInstaller.utils.hooks import collect_data_files
import googlemaps
from ortools.constraint_solver import routing_enums_pb2
from ortools.constraint_solver import pywrapcp
import pandas as pd
import requests
from rethinkdb import RethinkDB
jsonfile = []
r = RethinkDB()
depot = {
    'location': (22.35072200170769, 114.1243267585105)
}

API_KEY = 'AIzaSyDJeytG0koQtcGpU-5tWObFZBqPjYlDxd0'
url = "https://dev.virtualearth.net/REST/v1/Routes/DistanceMatrix?key=QJe42qLFnGh3YCD9D37D~ToZhwFtmV2Qz-7O5v6V9MA~ApnJDHL20D4XISksR_mE8_0YlgnKM2dpRqD24AsJaP2j21RpSt-vyCfR_3BMgFhl"
headers = {"Content-Type": "application/json; charset=utf-8",
           "Content-Length": 0}
data = {"origins": [{"latitude": 22.35072200170769, "longitude": 114.1243267585105}],
        "destinations": [], "travelMode": "driving"}

demands = [0]


def db():
    r.connect("localhost", 28015).repl()
    cursor = r.db("delivery").table("orders").run()
    orders = list(cursor)
    return orders


def coordinates(orders):

    for x in orders:
        if x['ReceiveTime'] == "":
            response = requests.get(url=(
                'https://maps.googleapis.com/maps/api/geocode/json?address='+x['RAddress']+'&key='+API_KEY))
            resp_json_payload = response.json()
            data["destinations"].append({
                "latitude": resp_json_payload["results"][0]["geometry"]["location"]["lat"], "longitude": resp_json_payload["results"][0]["geometry"]["location"]["lng"]
            })
        else:
            response = requests.get(url=(
                'https://maps.googleapis.com/maps/api/geocode/json?address='+x['SAddress']+'&key='+API_KEY))

            resp_json_payload = response.json()
            data["destinations"].append({
                "latitude": resp_json_payload["results"][0]["geometry"]["location"]["lat"], "longitude": resp_json_payload["results"][0]["geometry"]["location"]["lng"]
            })
        if x["Weight"] == 0:
            demands.append(2)
        else:
            demands.append(float(x["Weight"]))
    headers["Content-Length"] = str(len(json.dumps(data)))


googlemaps.Client(key=API_KEY)


def create_data_model(distance_matrix, time_matrix, num_vehicles):
    """Stores the data for the problem."""
    data = {}
    data['distance_matrix'] = distance_matrix
    data['num_vehicles'] = num_vehicles
    data['depot'] = 0
    data['demands'] = demands
    data['vehicle_capacities'] = [15, 15, 15, 15]
    data["time_matrix"] = time_matrix
    data['num_locations'] = len(data)
    return data


def build_distance_matrix(orderlist):
    w, h = len(orderlist)+1, len(orderlist)+1
    distance_matrix = [[None] * w for i in range(h)]
    time_matrix = [[None] * w for i in range(h)]
    response = requests.post(
        url="https://dev.virtualearth.net/REST/v1/Routes/DistanceMatrix?key=QJe42qLFnGh3YCD9D37D~ToZhwFtmV2Qz-7O5v6V9MA~ApnJDHL20D4XISksR_mE8_0YlgnKM2dpRqD24AsJaP2j21RpSt-vyCfR_3BMgFhl", headers=headers, json=data)
    resp_json_payload = response.json()
    for i in range(len(orderlist)+1):
        for j in range(len(orderlist)+1):
            if i == j:
                distance_matrix[i][j] = 0
                time_matrix[i][j] = 0
            else:
                distance_matrix[i][j] = math.ceil(resp_json_payload["resourceSets"][0]["resources"][0]['results'][j-1]["travelDistance"])
                time_matrix[i][j] = math.ceil(resp_json_payload["resourceSets"][0]["resources"][0]['results'][j-1]["travelDuration"])
    return distance_matrix, time_matrix


def print_solution(num_vehicles, model, manager, routing, solution):
    """Prints solution on console."""
    print(f'Objective: {solution.ObjectiveValue()}')
    total_distance = 0
    total_load = 0
    total_time = 0
    for vehicle_id in range(num_vehicles):
        dest=[]
        index = routing.Start(vehicle_id)
        plan_output = 'Route for vehicle {}:\n'.format(vehicle_id)
        route_distance = 0
        route_load = 0
        route_time = 0
        previous_node = 0
        while not routing.IsEnd(index):
            node_index = manager.IndexToNode(index)
            route_load += model['demands'][node_index]
            route_time += model["time_matrix"][previous_node][node_index]
            plan_output += ' {0} Load({1}) Time({2} min) -> '.format(
                node_index, route_load, route_time)
            previous_index = index
            previous_node = node_index
            index = solution.Value(routing.NextVar(index))
            route_distance += routing.GetArcCostForVehicle(
                previous_index, index, vehicle_id)
            dest.append(node_index)
        route_time += model["time_matrix"][previous_node][0]
        plan_output += ' {0} Load({1}) Time({2} min)\n'.format(manager.IndexToNode(index),route_load, route_time)
        plan_output += 'Distance of the route: {}m\n'.format(route_distance)
        plan_output += 'Load of the route: {}\n'.format(route_load)
        plan_output += 'Time of the route: {} min\n'.format(route_time)
        route={"vehicle_id":vehicle_id,"route":dest}
        jsonfile.append(route)
        print(plan_output)
        total_distance += route_distance
        total_load += route_load
        total_time += route_time

    print('Total distance of all routes: {}m'.format(total_distance))
    print('Total load of all routes: {}'.format(total_load))
    print('Total time of all routes: {} min'.format(total_time))


def extract_routes(num_vehicles, manager, routing, solution):
    routes = {}
    for vehicle_id in range(num_vehicles):
        routes[vehicle_id] = []
        index = routing.Start(vehicle_id)
        while not routing.IsEnd(index):
            routes[vehicle_id].append(manager.IndexToNode(index))
            previous_index = index
            index = solution.Value(routing.NextVar(index))
        routes[vehicle_id].append(manager.IndexToNode(index))
    return routes


def generate_solution(data, manager, routing, orderlist):
    """Solve the CVRP problem."""

    # Create and register a transit callback.
    def distance_callback(from_index, to_index):
        """Returns the distance between the two nodes."""
        # Convert from routing variable Index to distance matrix NodeIndex.
        from_node = manager.IndexToNode(from_index)
        to_node = manager.IndexToNode(to_index)
        return data['distance_matrix'][from_node][to_node]
    transit_callback_index = routing.RegisterTransitCallback(distance_callback)

    # Define cost of each arc.
    routing.SetArcCostEvaluatorOfAllVehicles(transit_callback_index)
    def demand_callback(from_index):
        """Returns the demand of the node."""
        # Convert from routing variable Index to demands NodeIndex.
        from_node = manager.IndexToNode(from_index)
        return data['demands'][from_node]
    demand_callback_index = routing.RegisterUnaryTransitCallback(
        demand_callback)



    # Add Capcity constraint.
    routing.AddDimensionWithVehicleCapacity(
        demand_callback_index,
        0,  # null capacity slack
        data['vehicle_capacities'],  # vehicle maximum capacities
        True,  # start cumul to zero
        'Capacity')

    # Add whole Time constraint
    time = 'Time'
    routing.AddDimension(
        transit_callback_index,
        30,  # allow waiting time
        240,  # maximum time per vehicle
        False,  # Don't force start cumul to zero.
        time)
    time_dimension = routing.GetDimensionOrDie(time)
    for i in range(data['num_vehicles']):
        routing.AddVariableMinimizedByFinalizer(
            time_dimension.CumulVar(routing.Start(i)))
        routing.AddVariableMinimizedByFinalizer(
            time_dimension.CumulVar(routing.End(i)))

        penalty = 1000
    for node in range(1, len(data['distance_matrix'])):
        if (orderlist[node-1]["Express"]):
            routing.AddDisjunction([manager.NodeToIndex(node)], penalty*5)
        else:
            routing.AddDisjunction([manager.NodeToIndex(node)], penalty)
    # Setting Local Search.
    search_parameters = pywrapcp.DefaultRoutingSearchParameters()
    search_parameters.local_search_metaheuristic = (
        routing_enums_pb2.LocalSearchMetaheuristic.AUTOMATIC)

    # Solve the problem.
    solution = routing.SolveWithParameters(search_parameters)
    return solution


def solve_vrp_for(time_matrix, distance_matrix, num_vehicles, orderlist):
    # Instantiate the data problem.
    model = create_data_model(distance_matrix, time_matrix, num_vehicles)

    # Create the routing index manager.
    distancemanager = pywrapcp.RoutingIndexManager(
        len(model['distance_matrix']), model['num_vehicles'], model['depot'])

    # Create Routing Model.
    routing = pywrapcp.RoutingModel(distancemanager)

    # Solve the problem
    solution = generate_solution(model, distancemanager, routing, orderlist)

    if solution:
        # Print solution on console.
        print_solution(num_vehicles, model, distancemanager,  routing, solution)
        routes = extract_routes(
            num_vehicles, distancemanager, routing, solution)
        return routes
    else:
        print('No solution found.')


def main():
    num_vehicles = 4
    orderlist = db()
    coordinates(orderlist)
    distancematrix, time_matrix = build_distance_matrix(orderlist)

    routes = solve_vrp_for(time_matrix, distancematrix,
                           num_vehicles, orderlist)

    jsonString = json.dumps(jsonfile)
    jsonFile = open("./data.json", "w")
    jsonFile.write(jsonString)
    jsonFile.close()

    return routes


if __name__ == "__main__":
    main()
