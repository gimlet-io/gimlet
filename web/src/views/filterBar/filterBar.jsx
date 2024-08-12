import { useState, useRef, useEffect } from 'react';
import { FunnelIcon, XMarkIcon } from '@heroicons/react/20/solid'

const FilterBar = (props) => {
  const { properties = [], filters, change } = props;

  const addFilter = (filter) => {
    change([...filters, filter]);
  }

  const filterValueByProperty = (property) => {
    const filter = filters.find(f => f.property === property)
    if (!filter) {
      return ""
    }

    return filter.value
  }

  const deleteFilter = (filter) => {
    change(filters.filter(f => f.property !== filter.property || f.value !== filter.value))
  }

  const resetFilters = () => {
    change([])
  }

  return (
    <div className="w-full">
      <div className="relative">
        <div className="absolute inset-y-0 left-0 flex items-center pl-3">
          <FunnelIcon className="filterIcon" aria-hidden="true" />
          {filters.map(filter => (
            <Filter key={filter.property+filter.value} filter={filter} deleteFilter={deleteFilter} />
          ))}
          <FilterInput properties={properties} filtersLength={filters.length} addFilter={addFilter} filterValueByProperty={filterValueByProperty} />
        </div>
        <div className="filter">
          &nbsp;
        </div>
        <div className="absolute inset-y-0 right-0 flex items-center p-1 px-2">
          <button onClick={resetFilters} className="py-1 px-2 bg-neutral-300 dark:bg-neutral-600 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-500 dark:text-neutral-400 rounded-full text-sm transition ease-in-out duration-150">Reset</button>
        </div>
      </div>
    </div>
  )
}

export default FilterBar;

const Filter = (props) => {
  const { filter } = props;
  return (
    <span className="py-0.5 ml-1 text-blue-50 dark:text-neutral-300 bg-blue-600 dark:bg-blue-900 rounded-full pl-3 pr-1" aria-hidden="true">
      <span>{filter.property}</span>{filter.property !== "Starred" && <span>: {filter.value}</span>}
      <span className="ml-1 px-1 bg-blue-400 dark:bg-blue-700 rounded-full ">
        <XMarkIcon className="cursor-pointer inline h-3 w-3" aria-hidden="true" onClick={() => props.deleteFilter(filter)}/>
      </span>
    </span>
  )
}

const FilterInput = (props) => {
  const [active, setActive] = useState(false)
  const [property, setProperty] = useState("")
  const [value, setValue] = useState("")
  const { properties, filtersLength, addFilter, filterValueByProperty } = props;
	const inputRef = useRef(null);

  const reset = () => {
    setActive(false)
    setProperty("")
    setValue("")
  }

  useEffect(() => {
    if (property !== "") {
      inputRef.current.focus();
    }  
  });

  return (
    <span className="relative w-48 ml-2">
      <span className="items-center flex">
        {property !== "" &&
          <span className='text-neutral-900 dark:text-neutral-200'>{property}: </span>
        }
        <input
          ref={inputRef}
          key={property}
          className={`${property ? "ml-10" : "" } block border-0 border-t border-b border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-800 pt-2 pb-1.5 px-1 text-neutral-900 dark:text-neutral-200 placeholder:text-neutral-300 dark:placeholder:text-neutral-600 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6`}
          placeholder={filtersLength !== 0 ? "" : "All repositories..."}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onFocus={() => {setActive(true)}}
          onBlur={() => {
            setTimeout(() => {
              setActive(false);
              if (value !== "") {
                if (property === "") {
                  addFilter({property: properties[0], value: value})
                } else {
                  addFilter({property, value})
                }
                reset()
              } else {
                if (property !== "") {
                  reset()
                }
              }
            }, 200);}
          }
          onKeyUp={(e) => {
            if (e.keyCode === 13){
              setActive(false)
              if (property === "") {
                addFilter({property: properties[0], value: value})
              } else {
                addFilter({property, value})
              }
              reset()
            }
            if (e.keyCode === 27){
              reset()
              // inputRef.current.blur();
            }
          }}
          type="search"
        />
      </span>
      {active && property === "" &&
      <div className="z-10 absolute bg-blue-100 w-48 p-2 text-blue-800">
        <ul className="">
          {properties.map(p => {
            if (filterValueByProperty(p) !== "") {
              return null;
            }

            return (
              <li
                key={p}
                className="cursor-pointer hover:bg-blue-200"
                onClick={() => {
                  if (p === "Starred") {
                    addFilter({property: p, value: "true"})
                    return
                  }

                  setProperty(p);
                  setActive(false);
                  }}>
                {p}
              </li>
          )})}
        </ul>
      </div>
      }
    </span>
  )
}
