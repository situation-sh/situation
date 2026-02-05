# Store

The **Store** is the internal memory where the agent put collected data. It has an internal structure with some helpers we advise to use. The store is closely related to the models defined in the `models` module. The store can be filled and queried as well (we try to make it **thread-safe**).

Basically, the store is a list of machines (`models.Machine`), so you may look for machines (to get information or to enrich it) or merely insert new ones.

<!-- prettier-ignore -->

!!! warning
The store and the models are rather instable. Especially, developers are likely to update the models (add/modify/remove attribute). Even if some functions help hiding the details, some changes can obviously have a wide impact.
