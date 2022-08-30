/*
 * Copyright (C) 2015, BMW Car IT GmbH
 *
 * Author: Sebastian Mattheis <sebastian.mattheis@bmw-carit.de>
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0 Unless required by applicable law or agreed to in
 * writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
 * language governing permissions and limitations under the License.
 */

package com.bmwcarit.barefoot.matcher;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import com.bmwcarit.barefoot.spatial.Geography;
import com.bmwcarit.barefoot.spatial.SpatialOperator;

import com.bmwcarit.barefoot.markov.KState;
import com.bmwcarit.barefoot.roadmap.Road;
import com.bmwcarit.barefoot.roadmap.RoadMap;
import com.esri.core.geometry.GeometryEngine;
import com.esri.core.geometry.Polyline;
import com.esri.core.geometry.WktExportFlags;

//import java.util.*

//import com.bmwcarit.barefoot.road.Heading;

/**
 * <i>k</i>-State data structure wrapper of {@link KState} for organizing state memory in HMM map
 * matching.
 */
public class MatcherKState extends KState<MatcherCandidate, MatcherTransition, MatcherSample> {
  private static final SpatialOperator spatial = new Geography();
    /**
     * Creates empty {@link MatcherKState} object with default parameters, which means capacity is
     * unbound.
     */
    public MatcherKState() {
        super();
    }

    /**
     * Creates a {@link MatcherKState} object from a JSON representation.
     *
     * @param json JSON representation of a {@link MatcherKState} object.
     * @param factory {@link MatcherFactory} for creation of matcher candidates and transitions.
     * @throws JSONException thrown on JSON extraction or parsing error.
     */
    public MatcherKState(JSONObject json, MatcherFactory factory) throws JSONException {
        super(json, factory);
    }

    /**
     * Creates an empty {@link MatcherKState} object and sets <i>&kappa;</i> and <i>&tau;</i>
     * parameters.
     *
     * @param k <i>&kappa;</i> parameter bounds the length of the state sequence to at most
     *        <i>&kappa;+1</i> states, if <i>&kappa; &ge; 0</i>.
     * @param t <i>&tau;</i> parameter bounds length of the state sequence to contain only states
     *        for the past <i>&tau;</i> milliseconds.
     */
    public MatcherKState(int k, long t) {
        super(k, t);
    }

    /**
     * Gets {@link JSONObject} with GeoJSON format of {@link MatcherKState} matched geometries.
     *
     * @return {@link JSONObject} with GeoJSON format of {@link MatcherKState} matched geometries.
     * @throws JSONException thrown on JSON extraction or parsing error.
     */
    public JSONObject toGeoJSON() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("type", "MultiLineString");
        JSONArray jsonsequence = new JSONArray();
        if (this.sequence() != null) {
            for (MatcherCandidate candidate : this.sequence()) {
                if (candidate.transition() == null) {
                    continue;
                }
                JSONObject jsoncandidate = new JSONObject(GeometryEngine
                        .geometryToGeoJson(candidate.transition().route().geometry()));
                jsonsequence.put(jsoncandidate.getJSONArray("coordinates"));
            }
        }
        json.put("coordinates", jsonsequence);
        return json;
    }

    /**
     * Gets JSON format String of {@link MatcherKState}, includes {@link JSONArray} String of
     * samples and {@link JSONArray} String of of matching results.
     *
     * @return JSON format String of {@link MatcherKState}, includes {@link JSONArray} String of
     *         samples and {@link JSONArray} String of of matching results.
     * @throws JSONException thrown on JSON extraction or parsing error.
     */
    public String toDebugJSON() throws JSONException { //add sa2, gcc
      JSONArray jsonpath = new JSONArray();
      if (this.sequence() != null) {
        long prior_time = this.samples().get(0).time() / 1000;
        long lastdt = prior_time;
        for (int i = 0; i < this.sequence().size(); ++i) {
            MatcherCandidate candidate = this.sequence().get(i);
            JSONObject jsoncandidate = new JSONObject();
            jsoncandidate.put("datetime", this.samples().get(i).time() / 1000);
            long time = this.samples().get(i).time() / 1000;
            long road = candidate.point().edge().refid();
            jsoncandidate.put("osm_id", Long.toString(road)); 
            long target = candidate.point().edge().target() ;
            jsoncandidate.put("target", target); 
            boolean forward = candidate.point().edge().forward() ;
            jsoncandidate.put("target", target); 
            double imputed_azimuth = candidate.point().azimuth();
            jsoncandidate.put("imputed_azimuth", imputed_azimuth);
            String sa2 = candidate.point().edge().sa2();
            jsoncandidate.put("sa2", sa2); 
            String gcc = candidate.point().edge().gcc();
            jsoncandidate.put("gcc", gcc); 

            
            if(i > 0){
              lastdt = this.samples().get(i-1).time() / 1000;
            }
            //jsoncandidate.put("frac", candidate.point().fraction());
            if (candidate.transition() != null) { //first obv will always be null so no logic checks for the prior_time related calculations
                if (lastdt != time){
                  prior_time = lastdt;
                   //only changes prior time when there has actually been a gap - accounts for dupe paths
                }
                double length = candidate.transition().route().length();
                if (length > 0) {
                  float gap = time - prior_time;
                  double imputed_speed = ((length/1000)/gap)*3600; //has to be specified in this form else it throws an infinite value
                  int n = candidate.transition().route().size();
                  double increment = (time-prior_time) / n;
                  JSONArray jsonroads = new JSONArray();

                  //actual matched road
                  JSONObject tupleMatched = new JSONObject();
                  tupleMatched.put("osm_id", Long.toString(road) );
                  tupleMatched.put("sa2",sa2);
                  tupleMatched.put("gcc", gcc) ;
                  tupleMatched.put("forward", forward);
                  tupleMatched.put("imputed_azimuth", imputed_azimuth);
                  tupleMatched.put("target", Long.toString(target));
                  tupleMatched.put("datetime", time);
                  
                  //tupleMatched.put("n_legs", n);
                  //duplicate json fields for backwards compatability
                  tupleMatched.put("type", "matched path");
                  tupleMatched.put("length", length);
                  //adding to parent
                  jsoncandidate.put("type", "matched path");
                  jsoncandidate.put("length", length);
                  jsoncandidate.put("source_frac", 1-candidate.transition().route().source().fraction());
                  //checkthis
                  jsoncandidate.put("target_frac", candidate.transition().route().target().fraction());
                  jsoncandidate.put("source_id", Long.toString(candidate.transition().route().get(0).refid()));
                  //jsoncandidate.put("target_id", candidate.transition().route().get(n-1).refid());
                  //tupleMatched.put("gap", gap);
                  if (imputed_speed < 120){
                    jsoncandidate.put("imputed_speed" , imputed_speed);
                  }
                  jsonroads.put(tupleMatched);

                  for (int j = 1; j < n; ++j) { //starting at 1 because first imputed obvs is last recorded obv
                      JSONObject tuple = new JSONObject();
                      tuple.put("osm_id", Long.toString(candidate.transition().route().get(j).refid()));
                      tuple.put("target", Long.toString(candidate.transition().route().get(j).target()));
                      tuple.put("forward", candidate.transition().route().get(j).forward());
                      tuple.put("imputed_azimuth", candidate.transition().route().get(j).med_azi());
                      tuple.put("sa2", candidate.transition().route().get(j).sa2());
                      tuple.put("gcc", candidate.transition().route().get(j).gcc());
                      //tuple.put("imputed_azimuth", spatial.azimuth(candidate.transition().route().get(j).geometry(), 0.5));


                      // tuple.put(
                      //         "geom",
                      //         GeometryEngine.geometryToWkt(candidate.transition().route().get(j)
                      //                 .geometry(), WktExportFlags.wktExportLineString));
                      if (imputed_speed < 140){
                        tuple.put("imputed_speed" ,imputed_speed);
                      }
                      //tuple.put("datetime", time);
                      //tuple.put("n_legs", n);
                      //tuple.put("length", length);
                      //tuple.put("gap", gap);
                      double imputed_time = prior_time + (j) * increment;
                      tuple.put("imputed_time", imputed_time);
                      tuple.put("datetime", time);


                      tuple.put("type", "imputed");
                      jsonroads.put(tuple);
                  }
                  jsoncandidate.put("roads", jsonroads);
                  jsonpath.put(jsoncandidate);
              }
          } else {
                jsoncandidate.put("type", "matched no path");
                //MatcherCandidate lastcandidate = this.sequence().get(i);
                
                jsonpath.put(jsoncandidate);
          }
          
        }
      }
      return jsonpath.toString();
          }

    /**
     * Gets {@link JSONArray} of {@link MatcherKState} with map matched positions, represented by
     * road id and fraction, and the geometry of the routes.
     *
     * @return {@link JSONArray} of {@link MatcherKState} with map matched positions, represented by
     *         road id and fraction, and the geometry of the routes.
     * @throws JSONException thrown on JSON extraction or parsing error.
     */
    public JSONArray toSlimJSON() throws JSONException {
        JSONArray json = new JSONArray();
        if (this.sequence() != null) {
            for (MatcherCandidate candidate : this.sequence()) {
                JSONObject jsoncandidate = candidate.point().toJSON();
                if (candidate.transition() != null) {
                    jsoncandidate.put("route",
                            GeometryEngine.geometryToWkt(candidate.transition().route().geometry(),
                                    WktExportFlags.wktExportLineString));
                }
                json.put(jsoncandidate);
            }
        }
        return json;
    }

    private Polyline monitorRoute(MatcherCandidate candidate) {
        Polyline routes = new Polyline();
        MatcherCandidate predecessor = candidate;
        while (predecessor != null) {
            MatcherTransition transition = predecessor.transition();
            if (transition != null) {
                Polyline route = transition.route().geometry();
                routes.startPath(route.getPoint(0));
                for (int i = 1; i < route.getPointCount(); ++i) {
                    routes.lineTo(route.getPoint(i));
                }
            }
            predecessor = predecessor.predecessor();
        }
        return routes;
    }

    public JSONObject toMonitorJSON() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("time", sample().time());
        json.put("point", GeometryEngine.geometryToWkt(estimate().point().geometry(),
                WktExportFlags.wktExportPoint));
        Polyline routes = monitorRoute(estimate());
        if (routes.getPathCount() > 0) {
            json.put("route",
                    GeometryEngine.geometryToWkt(routes, WktExportFlags.wktExportMultiLineString));
        }

        JSONArray candidates = new JSONArray();
        for (MatcherCandidate candidate : vector()) {
            JSONObject jsoncandidate = new JSONObject();
            jsoncandidate.put("point", GeometryEngine.geometryToWkt(candidate.point().geometry(),
                    WktExportFlags.wktExportPoint));
            jsoncandidate.put("prob",
                    Double.isInfinite(candidate.filtprob()) ? "Infinity" : candidate.filtprob());

            routes = monitorRoute(candidate);
            if (routes.getPathCount() > 0) {
                jsoncandidate.put("route", GeometryEngine.geometryToWkt(routes,
                        WktExportFlags.wktExportMultiLineString));
            }
            candidates.put(jsoncandidate);
        }
        json.put("candidates", candidates);
        return json;
    }

    private static String getOSMRoad(Road road) {
        return road.base().refid() + ":" + road.source() + ":" + road.target();
    }

    public JSONObject toOSMJSON(RoadMap map) throws JSONException {
        JSONObject json = this.toJSON();

        if (json.has("candidates")) {
            JSONArray candidates = json.getJSONArray("candidates");

            for (int i = 0; i < candidates.length(); ++i) {
                {
                    JSONObject point = candidates.getJSONObject(i).getJSONObject("candidate")
                            .getJSONObject("point");
                    Road road = map.get(point.getLong("osm_id"));
                    if (road == null) {
                        throw new JSONException("road not found in map");
                    }
                    point.put("osm_id", getOSMRoad(road));
                }
                if (candidates.getJSONObject(i).getJSONObject("candidate").has("transition")) {
                    JSONObject route = candidates.getJSONObject(i).getJSONObject("candidate")
                            .getJSONObject("transition").getJSONObject("route");

                    JSONArray roads = route.getJSONArray("roads");
                    JSONArray osmroads = new JSONArray();
                    for (int j = 0; j < roads.length(); ++j) {
                        Road road = map.get(roads.getLong(j));
                        osmroads.put(getOSMRoad(road));
                    }
                    route.put("roads", osmroads);

                    {
                        JSONObject source = route.getJSONObject("source");
                        Road road = map.get(source.getLong("osm_id"));
                        source.put("osm_id", getOSMRoad(road));
                    }
                    {
                        JSONObject target = route.getJSONObject("target");
                        Road road = map.get(target.getLong("osm_id"));
                        target.put("osm_id", getOSMRoad(road));
                    }
                }
            }
        }
        return json;
    }
}
