FROM osrm/osrm-backend
RUN sed -i 's/handlers = Sequence /result.name = way:id()\nhandlers = Sequence /g' /opt/car.lua && \
	sed -i 's/WayHandlers.names/--WayHandlers.names/g' /opt/car.lua
