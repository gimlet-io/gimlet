import React from "react";
import {StarIcon} from "@heroicons/react/outline";
import {StarIcon as SolidStarIcon} from "@heroicons/react/solid";

function RepoCard(props) {
  const {name, services, navigateToRepo, favorite, favoriteHandler} = props;
  const numberOfServicesOnCard = 5;
  const truncatedServices = services.length > numberOfServicesOnCard ? services.slice(0, numberOfServicesOnCard) : services;

  const serviceWidgets = truncatedServices.map(service => {
    let ingressWidgets = [];
    if (service.ingresses !== undefined) {
      ingressWidgets = service.ingresses.map(ingress => (
        <p className="text-gray-400">
          <span>{ingress.url}</span>
          <a href={`https://${ingress.url}`} target="_blank" rel="noopener noreferrer"
             onClick={(e) => {
               e.stopPropagation();
               return true
             }}>
            <svg xmlns="http://www.w3.org/2000/svg"
                 className="inline fill-current hover:text-teal-300 ml-1 text-gray-500 hover:text-gray-700" width="12"
                 height="12" viewBox="0 0 24 24">
              <path d="M0 0h24v24H0z" fill="none"/>
              <path
                d="M19 19H5V5h7V3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2v-7h-2v7zM14 3v2h3.59l-9.83 9.83 1.41 1.41L19 6.41V10h2V3h-7z"/>
            </svg>
          </a>
        </p>
      ))
    }

    return (
      <div>
        <p className="text-xs">{service.service.namespace}/{service.service.name}
          <span
            className="flex-shrink-0 inline-block px-2 py-0.5 mx-1 text-green-800 text-xs font-medium bg-green-100 rounded-full">
          {service.env}
        </span>
        </p>
        <ul className="text-xs pl-2">
          {ingressWidgets}
        </ul>
      </div>
    )
  })

  return (
    <div className="relative">
      <div className="absolute top-0 right-0 m-6">
        {favorite &&
        <SolidStarIcon
          className="h-4 w-4 text-green-400 hover:text-gray-400 cursor-pointer"
          onClick={() => favoriteHandler(name)}
        />
        }
        {!favorite &&
        <StarIcon
          className="h-4 w-4 text-gray-400 hover:text-green-400 cursor-pointer"
          onClick={() => favoriteHandler(name)}
        />
        }
      </div>
      <div className="w-full flex items-center justify-between p-6 space-x-6 cursor-pointer"
           onClick={() => navigateToRepo(name)}>
        <div className="flex-1 truncate">
          <p className="text-sm font-bold">{name}</p>
          <div className="p-2 space-y-2">
            {serviceWidgets}
            {services.length > numberOfServicesOnCard ?  <p>...</p> : null}
          </div>
        </div>
      </div>
    </div>
  )
}

export default RepoCard;
